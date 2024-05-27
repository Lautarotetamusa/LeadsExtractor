package pkg

import (
	"encoding/json"
	"fmt"
	"leadsextractor/infobip"
	"leadsextractor/models"
	"log"
	"net/http"
	"time"
)

func (s *Server) HandlePipedriveOAuth(w http.ResponseWriter, r *http.Request) error {
	code := r.URL.Query().Get("code")
	log.Println("Code:", code)
	s.pipedrive.ExchangeCodeToToken(code)
	w.Write([]byte(s.pipedrive.State))
	return nil
}

func (s *Server) GetCommunications(w http.ResponseWriter, r *http.Request) error {
    dateString := r.URL.Query().Get("date")
    if dateString == "" {
        return fmt.Errorf("el parametro date es obligatorio")
    }
    const format = "01-02-2006"

    date, err := time.Parse(format, dateString)
    if err != nil{
        return err
    }

    query := "CALL getCommunications(?)"

	communications := []models.CommunicationDB{}
	if err := s.Store.Db.Select(&communications, query, date); err != nil {
        log.Fatal(err)
	}
    s.logger.Info(fmt.Sprintf("Se encontraron %d comunicaciones", len(communications)))

	w.Header().Set("Content-Type", "application/json")
	res := struct {
		Success bool        `json:"success"`
        Length  int         `json:"length"`
		Data    interface{} `json:"data"`
	}{true, len(communications), &communications}

	json.NewEncoder(w).Encode(res)
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

    s.Store.InsertCommunication(c, lead, source, isNewLead)

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
