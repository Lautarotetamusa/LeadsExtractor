package pkg

import (
	"encoding/json"
	"fmt"
	"leadsextractor/models"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type GetFlowPayload struct{
    Uuid uuid.NullUUID `json:"uuid"`
}

func (s *Server) GetFlows(w http.ResponseWriter, r *http.Request) error {
    dataResponse(w, config.Flows)
    return nil
}

func NewAction(w http.ResponseWriter, r *http.Request) error {
    var rules []Rule
    if err := json.NewDecoder(r.Body).Decode(&rules); err != nil {
        return err
    }

    uuid, err := uuid.NewRandom()
    if err != nil {
        return fmt.Errorf("no se pudo generar una uuid: %s", err)
    }
    config.Flows[uuid] = rules
    saveConfig("actions.json")

    dataResponse(w, rules)
    return nil
}

func (s *Server) SetFlowAsMain(w http.ResponseWriter, r *http.Request) error {
    var body GetFlowPayload
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return err
	}

    if _, exists := config.Flows[body.Uuid.UUID]; !exists {
        return fmt.Errorf("no existe ningun flow con id %s", body.Uuid.UUID)
    }
   
    config.Main = body.Uuid.UUID
    saveConfig("actions.json")

    return nil
}

func (s *Server) NewBroadcast(w http.ResponseWriter, r *http.Request) error {
    var body GetFlowPayload
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
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

