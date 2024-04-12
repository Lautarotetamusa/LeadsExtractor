package pkg

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"leadsextractor/infobip"
	"leadsextractor/models"
	"leadsextractor/store"
	"net/http"
	"slices"

	"github.com/jmoiron/sqlx"
)

func (s *Server) HandleListCommunication(w http.ResponseWriter, r *http.Request) error{
    query := `
    SELECT 
        C.created_at as "Fecha extraccion", C.lead_date as "Fecha lead", A.name as "Asesor asignado", S.type, P.portal, L.name, C.url, L.phone, L.email,
        P.*,
        C.zones, C.mt2_terrain, C.mt2_builded, C.baths, C.rooms
    FROM Communication C
    INNER JOIN Leads L 
        ON C.lead_phone = L.phone
    INNER JOIN Source S
        ON C.source_id = S.id
    INNER JOIN Asesor A
        ON L.asesor = A.phone
    LEFT JOIN Property P
        ON S.property_id = P.id;`;

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
    row := make([]string, colCount)
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
    fmt.Printf("communication: %+v\n", c)

    source, err := getSource(s.db, c)
    if err != nil{
        return err
    }

    lead, isNewLead, err := insertOrGetLead(s.db, s.roundRobin, c)
    if err != nil{
        return err
    }
    c.Asesor = lead.Asesor
    
    query := `INSERT INTO Communication(lead_phone, source_id, lead_date, url, zones, mt2_terrain, mt2_builded, baths, rooms) 
    VALUES (:lead_phone, :source_id, :lead_date, :url, :zones, :mt2_terrain, :mt2_builded, :baths, :rooms)`
    _, err = s.db.NamedExec(query, map[string]interface{}{
        "lead_phone": lead.Phone,
        "source_id": source.Id,
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

    /*
    if isNewLead {
        formatMsg(lead, bienvenida2)
    
        s.infobipApi.SendWppMedia(lead.Phone, imageId)
        s.infobipApi.SendWppText(lead.Phone, bienvenida1)
        s.infobipApi.SendWppText(lead.Phone, bienvenida2)
        s.infobipApi.SendWppMedia(lead.Phone, videoId)
    }else{
        s.infobipApi.SendWppTemplate(lead.Phone, "2do_mensaje_bienvenida")
    }

    s.infobipApi.SendWppTemplate(lead.Asesor.Phone, "msg_asesor")
    */

    fmt.Printf("lead: %#v\n", lead)

    infobipLead := infobip.Communication2Infobip(&c) 
    s.infobipApi.SaveLead(infobipLead)
    s.infobipApi.SendWppTextMessage(lead.Phone, "hola amigo")

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
    validSources := []string{"whatsapp", "ivr", "inmuebles24", "lamudi", "casasyterrenos", "propiedades"}
    if !slices.Contains(validSources, c.Fuente){
        return nil, fmt.Errorf("La fuente %s es incorrecta, debe ser (whatsapp, ivr, inmuebles24, lamudi, casasyterrenos, propiedades)", c.Fuente)
    }

    if c.Fuente == "whatsapp" || c.Fuente == "ivr"{
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
