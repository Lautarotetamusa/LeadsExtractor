package pkg

import (
	"encoding/json"
	"leadsextractor/infobip"
	"leadsextractor/models"
	"log"
	"net/http"
)

func (s *Server) HandlePipedriveOAuth(w http.ResponseWriter, r *http.Request) error {
	code := r.URL.Query().Get("code")
	log.Println("Code:", code)
	s.pipedrive.ExchangeCodeToToken(code)
	w.Write([]byte(s.pipedrive.State))
	return nil
}

func (s *Server) HandleListCommunication(w http.ResponseWriter, r *http.Request) error {
	query := "CALL communicationList()"

	rows, err := s.Store.Db.Queryx(query)
	if err != nil {
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
	res := struct {
		Success   bool        `json:"success"`
		Headers   []string    `json:"headers"`
		RowsCount int         `json:"row_count"`
		ColsCount int         `json:"col_count"`
		Rows      interface{} `json:"rows"`
	}{true, cols, rowCount, colCount, result}
	json.NewEncoder(w).Encode(res)
	return nil
}

func (s *Server) AssignAsesor(w http.ResponseWriter, r *http.Request) error {
	c := &models.Communication{}
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(c)
	if err != nil {
		return err
	}

	lead, isNewLead, err := s.Store.InsertOrGetLead(s.roundRobin, c)
	if err != nil {
		return err
	}
	c.Asesor = lead.Asesor

	w.Header().Set("Content-Type", "application/json")
	res := struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
		IsNew   bool        `json:"is_new"`
	}{true, c, isNewLead}

	json.NewEncoder(w).Encode(res)
	return nil
}

func (s *Server) HandleNewCommunication(w http.ResponseWriter, r *http.Request) error {
	c := &models.Communication{}
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(c)
	if err != nil {
		return err
	}

	source, err := s.Store.GetSource(c)
	if err != nil {
		return err
	}

	lead, isNewLead, err := s.Store.InsertOrGetLead(s.roundRobin, c)
	if err != nil {
		return err
	}
	c.Asesor = lead.Asesor

	query := `INSERT INTO Communication(lead_phone, source_id, new_lead, lead_date, url, zones, mt2_terrain, mt2_builded, baths, rooms) 
    VALUES (:lead_phone, :source_id, :new_lead, :lead_date, :url, :zones, :mt2_terrain, :mt2_builded, :baths, :rooms)`
	_, err = s.Store.Db.NamedExec(query, map[string]interface{}{
		"lead_phone":  lead.Phone,
		"source_id":   source.Id,
		"new_lead":    isNewLead,
		"lead_date":   c.FechaLead,
		"url":         c.Link,
		"zones":       c.Busquedas.Zonas,
		"mt2_terrain": c.Busquedas.TotalArea,
		"mt2_builded": c.Busquedas.CoveredArea,
		"baths":       c.Busquedas.Banios,
		"rooms":       c.Busquedas.Recamaras,
	})
	if err != nil {
		return err
	}

	go s.pipedrive.SaveCommunication(c)

	infobipLead := infobip.Communication2Infobip(c)
	s.infobipApi.SaveLead(infobipLead)

	w.Header().Set("Content-Type", "application/json")
	res := struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
		IsNew   bool        `json:"is_new"`
	}{true, c, isNewLead}

	json.NewEncoder(w).Encode(res)
	return nil
}
