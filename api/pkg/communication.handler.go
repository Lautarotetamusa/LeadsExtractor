package pkg

import (
	"encoding/json"
	"fmt"
	"leadsextractor/models"
	"log"
	"net/http"
	"time"
)

func (s *Server) GetCommunications(w http.ResponseWriter, r *http.Request) error {
    dateString := r.URL.Query().Get("date")
    if dateString == "" {
        return fmt.Errorf("el parametro date es obligatorio")
    }
    const format = "01-02-2006"

    var isNew *bool;
    isNewStr := r.URL.Query().Get("is_new")

    switch isNewStr {
    case "":
        isNew = nil
    case "false":
        val := false
        isNew = &val
    case "true":
        val := true
        isNew = &val
    default:
        return fmt.Errorf("el parametro is_new tiene que ser un booleano")
    }

    date, err := time.Parse(format, dateString)
    if err != nil{
        return err
    }

    query := "CALL getCommunications(?, ?)"

	communications := []models.Communication{}
	if err := s.Store.Db.Select(&communications, query, date, isNew); err != nil {
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


func (s *Server) NewCommunication(w http.ResponseWriter, r *http.Request) error {
	c := &models.Communication{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(c); err != nil {
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

    go s.flowHandler.manager.RunMainFlow(c)
        
    if err = s.Store.InsertCommunication(c, lead, source); err != nil {
        s.logger.Error(err.Error(), "path", "InsertCommunication")
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
