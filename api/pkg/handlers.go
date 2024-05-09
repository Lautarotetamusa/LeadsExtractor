package pkg

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"leadsextractor/infobip"
	"leadsextractor/models"
	"leadsextractor/store"
	"log"
	"net/http"
	"slices"

	"github.com/jmoiron/sqlx"
)

func (s *Server) HandlePipedriveOAuth(w http.ResponseWriter, r *http.Request) error{
    code := r.URL.Query().Get("code")
    log.Println("Code:", code)
    s.pipedrive.ExchangeCodeToToken(code)
    w.Write([]byte(s.pipedrive.State))
    return nil 
}

func (s *Server) HandleListCommunication(w http.ResponseWriter, r *http.Request) error{
    query := "CALL communicationList()";

    rows, err := s.db.Queryx(query)
    if err != nil{
        return err
    }    
    cols, err := rows.Columns()
    if err != nil {
        return err
    }
    colCount := len(cols)

    // Result is your slice string.
    rawResult := make([][]byte, colCount)
    dest := make([]interface{}, colCount) // A temporary interface{} slice
    var result [][]string

    for i := range rawResult {
        dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
    }

    rowCount := 0
    for rows.Next() {
        rowCount += 1
        err = rows.Scan(dest...)
        if err != nil {
            return err
        }

        //Lo hago aca dentro de nuevo porque el append apendea un puntero
        //Es decir si haces a = append(a, b) y despues  cambias b, a va a tener el nuevo valor de b
        row := make([]string, colCount)
        for i, raw := range rawResult {
            row[i] = string(raw)
        }
        result = append(result, row)
    }

    w.Header().Set("Content-Type", "application/json")
    res := struct{
        Success bool `json:"success"`
        Headers []string `json:"headers"`
        RowsCount int `json:"row_count"` 
        ColsCount int `json:"col_count"` 
        Rows interface{} `json:"rows":`
    }{true, cols, rowCount, colCount, result}
    json.NewEncoder(w).Encode(res)
    return nil
}

func (s *Server) HandleNewCommunication(w http.ResponseWriter, r *http.Request) error{
    c := models.Communication{}
    reqBody, err := ioutil.ReadAll(r.Body)
    err = json.Unmarshal(reqBody, &c)
    if err != nil{
        return err
    }

    source, err := getSource(s.db, c)
    if err != nil{
        return err
    }

    lead, isNewLead, err := insertOrGetLead(s.db, s.roundRobin, c)
    if err != nil{
        return err
    }
    c.Asesor = lead.Asesor
    
    query := `INSERT INTO Communication(lead_phone, source_id, new_lead, lead_date, url, zones, mt2_terrain, mt2_builded, baths, rooms) 
    VALUES (:lead_phone, :source_id, :new_lead, :lead_date, :url, :zones, :mt2_terrain, :mt2_builded, :baths, :rooms)`
    _, err = s.db.NamedExec(query, map[string]interface{}{
        "lead_phone": lead.Phone,
        "source_id": source.Id,
        "new_lead": isNewLead, 
        "lead_date": c.FechaLead, 
        "url": c.Link,
        "zones": c.Busquedas.Zonas,
        "mt2_terrain": c.Busquedas.TotalArea,
        "mt2_builded": c.Busquedas.CoveredArea,
        "baths": c.Busquedas.Banios,
        "rooms": c.Busquedas.Recamaras,
    })
    if err != nil{
        return err
    }

    fmt.Printf("%v\n", c)
    go s.pipedrive.SaveCommunication(&c)

    infobipLead := infobip.Communication2Infobip(&c) 
    s.infobipApi.SaveLead(infobipLead)

    //Return
    w.Header().Set("Content-Type", "application/json")
    res := struct{
        Success bool `json:"success"`
        Data interface{} `json:"data":`
        IsNew bool `json:"is_new"`
    }{true, c, isNewLead}
    json.NewEncoder(w).Encode(res)
    return nil
}

func getSource(db *sqlx.DB, c models.Communication) (*models.Source, error){
    source := models.Source{}
    validSources := []string{"whatsapp", "ivr", "viewphone", "inmuebles24", "lamudi", "casasyterrenos", "propiedades"}
    if !slices.Contains(validSources, c.Fuente){
        return nil, fmt.Errorf("La fuente %s es incorrecta, debe ser (whatsapp, ivr, inmuebles24, lamudi, casasyterrenos, propiedades)", c.Fuente)
    }

    if c.Fuente == "whatsapp" || c.Fuente == "ivr" || c.Fuente == "viewphone"{
        err := db.Get(&source, "SELECT * FROM Source WHERE type LIKE ?", c.Fuente) 
        if err != nil{
            return nil, fmt.Errorf("Source: %s no existe", c.Fuente)
        }
        return &source, nil
    }

    property, err := insertOrGetProperty(db, c)
    if err != nil{
        return nil, err
    }

    err = db.Get(&source, "SELECT * FROM Source WHERE property_id=?", property.Id)
    if err != nil{
        return nil, err
    }
    return &source, nil
}

func insertOrGetProperty(db *sqlx.DB, c models.Communication) (*models.Property, error){
    property := models.Property{}
    query :=  "SELECT * FROM Property WHERE portal_id LIKE ? AND portal = ? LIMIT 1"
    err := db.Get(&property, query, c.Propiedad.ID, c.Fuente)

    if err == sql.ErrNoRows {
        fmt.Println("No se encontro Property")

        query := "INSERT INTO Property (portal_id, title, url, price, ubication, tipo, portal) VALUES (:portal_id, :title, :url, :price, :ubication, :tipo, :portal)"
        property = models.Property{
            PortalId: c.Propiedad.ID,
            Title: c.Propiedad.Titulo,
            Url: c.Propiedad.Link,
            Price: c.Propiedad.Precio,
            Ubication: c.Propiedad.Ubicacion,
            Tipo: c.Propiedad.Tipo,
            Portal: c.Fuente,
        }
        if _, err := db.NamedExec(query, &property); err != nil {
            return nil, err
        }

        query = "SELECT * FROM Property WHERE portal_id LIKE ? AND portal = ? LIMIT 1"
        err := db.Get(&property, query, c.Propiedad.ID, c.Fuente)

        //Cargamos el source
        source := models.Source{
            Tipo: "property",
            PropertyId: sql.NullInt16{Int16: int16(property.Id), Valid: true},
        }
        query = "INSERT INTO Source (type, property_id) VALUES(:type, :property_id)"
        if _, err = db.NamedExec(query, source); err != nil{
            return nil, err
        }
    }

    return &property, nil
}

func insertOrGetLead(db *sqlx.DB, rr *RoundRobin, c models.Communication) (*models.Lead, bool, error){
    var isNewLead = false
    var lead *models.Lead

    leadStorage := store.LeadMySqlStorage{Db: db,}
    lead, err := leadStorage.GetOne(c.Telefono)

    if err == sql.ErrNoRows{
        isNewLead = true
        c.Asesor = rr.next()

        lead, err = leadStorage.Insert(&models.CreateLead{
            Name: c.Nombre,
            Phone: c.Telefono,
            Email: sql.NullString{String: c.Email},
            AsesorPhone: c.Asesor.Phone,
        })

        if err != nil{
            return nil, isNewLead, err
        }
    }else if err != nil{
        return nil, isNewLead, err
    }

    return lead, isNewLead, nil
}
