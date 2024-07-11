package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"leadsextractor/flow"
	"leadsextractor/models"
	"leadsextractor/store"
	"net/http"
	"text/template"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type GetFlowPayload struct{
    Uuid uuid.NullUUID `json:"uuid"`
}

type FlowHandler struct {
    manager *flow.FlowManager
}

func NewFlowHandler(m *flow.FlowManager) *FlowHandler {
    return &FlowHandler{
        manager: m,
    }
}

func (h *FlowHandler) GetFlows(w http.ResponseWriter, r *http.Request) error {
    dataResponse(w, h.manager.Flows)
    return nil
}

func (h *FlowHandler) NewFlow(w http.ResponseWriter, r *http.Request) error {
    var f flow.Flow
    if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
        return err
    }

    if err := h.manager.AddFlow(&f); err != nil {
        return err
    }

    dataResponse(w, f)
    return nil
}

func (h *FlowHandler) DeleteFlow(w http.ResponseWriter, r *http.Request) error {
    uuid, err := h.getUUIDFromParam(r)
    if err != nil {
        return err
    }
    if err := h.manager.DeleteFlow(*uuid); err != nil {
        return err
    }

    dataResponse(w, h.manager.Flows)
    return nil
}

func (h *FlowHandler) UpdateFlow(w http.ResponseWriter, r *http.Request) error {
    var f flow.Flow
    if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
        return err
    }

    uuid, err := h.getUUIDFromParam(r)
    if err != nil {
        return err
    }

    if err := h.manager.UpdateFlow(&f, *uuid); err != nil {
        return err
    }

    dataResponse(w, h.manager.Flows[*uuid])
    return nil
}

func (h *FlowHandler) SetFlowAsMain(w http.ResponseWriter, r *http.Request) error {
    uuid, err := h.getUUIDFromBody(r)
    if err != nil {
        return err
    }
    h.manager.SetMain(*uuid)

    dataResponse(w, h.manager.Flows[*uuid])
    return nil
}

func (s *Server) NewBroadcast(w http.ResponseWriter, r *http.Request) error {
    uuid, err := s.flowHandler.getUUIDFromBody(r)
    if err != nil {
        return err
    }
   
    params := store.QueryParam{
        IsNew: true,
    }
	communications, err := s.Store.GetCommunications(&params)
	if err != nil {
        return err
	}

    go s.flowHandler.manager.Broadcast(communications, *uuid)

	w.Header().Set("Content-Type", "application/json")
	res := struct {
		Success bool    `json:"success"`
		Count   int     `json:"count"`
	}{true, len(communications)}

	json.NewEncoder(w).Encode(res)
	return nil
}

func (h *FlowHandler) getUUIDFromBody(r *http.Request) (*uuid.UUID, error) {
    var body GetFlowPayload
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

    if _, exists := h.manager.Flows[body.Uuid.UUID]; !exists {
        return nil, fmt.Errorf("no existe ningun flow con id %s", body.Uuid.UUID)
    }

    return &body.Uuid.UUID, nil
}

func (h *FlowHandler) getUUIDFromParam(r *http.Request) (*uuid.UUID, error) {
    uuidStr, exists := mux.Vars(r)["uuid"]
    if !exists {
        return nil, fmt.Errorf("se necesita pasar un uuid")
    }
    uuid, err := uuid.Parse(uuidStr)
    if err != nil {
        return nil, err
    }
    if _, exists := h.manager.Flows[uuid]; !exists {
        return nil, fmt.Errorf("no existe ningun flow con id %s", uuid)
    }

    return &uuid, nil
}


func FormatMsg(tmpl string, c *models.Communication) string {
    t := template.Must(template.New("txt").Parse(tmpl))
    buf := &bytes.Buffer{}
    if err := t.Execute(buf, c); err != nil {
        return ""
    }
    return buf.String()
}

