package pkg

import (
	"encoding/json"
	"fmt"
	"leadsextractor/models"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

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

	communications := []models.Communication{}
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

func (s *Server) NewBroadcast(w http.ResponseWriter, r *http.Request) error {
    type payload struct{
        Uuid uuid.NullUUID `json:"uuid"`
    }
    var body payload

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return err
	}
    
    query := "CALL getCommunications(DATE_SUB(NOW(), interval 20 day))"
	communications := []models.Communication{}
	if err := s.Store.Db.Select(&communications, query); err != nil {
        log.Fatal(err)
	}

    broadcast(communications, body.Uuid.UUID)
	w.Header().Set("Content-Type", "application/json")
	res := struct {
		Success bool    `json:"success"`
		Count   int     `json:"count"`
	}{true, len(communications)}

	json.NewEncoder(w).Encode(res)
	return nil
}

func (s *Server) NewCommunication(w http.ResponseWriter, r *http.Request) error {
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
    fmt.Printf("%v\n", source)

	lead, err := s.Store.InsertOrGetLead(s.roundRobin, c)
	if err != nil {
		return err
	}
	c.Asesor = lead.Asesor

    //go runFlow(c)
        
    err = s.Store.InsertCommunication(c, lead, source)
    if err != nil {
        return err
    }

	w.Header().Set("Content-Type", "application/json")
	res := struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
		IsNew   bool        `json:"is_new"`
	}{true, c, c.IsNew}

	json.NewEncoder(w).Encode(res)
	return nil
}
