package pkg

import (
	"encoding/json"
	"fmt"
	"leadsextractor/flow"
	"leadsextractor/store"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type GetFlowPayload struct {
    Uuid uuid.NullUUID `json:"uuid"`
}

type NewBroadcastPayload struct {
    Uuid        uuid.UUID           `json:"uuid"`
    Condition   store.QueryParam    `json:"condition"`
}

type FlowHandler struct {
    manager *flow.FlowManager
}

type FlowResponse struct {
    IsMain  bool        `json:"is_main"`
    Rules   []flow.Rule `json:"rules"`
    Uuid    uuid.UUID   `json:"uuid"`
}

func NewFlowHandler(m *flow.FlowManager) *FlowHandler {
    return &FlowHandler{
        manager: m,
    }
}
func (h *FlowHandler) GetConfig(w http.ResponseWriter, r *http.Request) error {
    actions := h.manager.GetActions()
    dataResponse(w, actions)
    return nil
}

func (h *FlowHandler) GetFlows(w http.ResponseWriter, r *http.Request) error {
    flows := h.manager.GetFlows()
    dataResponse(w, h.parseFlows(flows))
    return nil
}

func (h *FlowHandler) GetMainFlow(w http.ResponseWriter, r *http.Request) error {
    flow, err := h.manager.GetMainFlow()
    if err != nil {
        return err
    }

    uuid := h.manager.GetMain()
    dataResponse(w, h.parseFlow(flow, &uuid))
    return nil
}

func (h *FlowHandler) GetFlow(w http.ResponseWriter, r *http.Request) error {
    uuid, err := h.getUUIDFromParam(r)
    if err != nil {
        return err
    }
    flow, err := h.manager.GetFlow(*uuid)
    if err != nil {
        return err
    }

    dataResponse(w, h.parseFlow(flow, uuid))
    return nil
}

func (h *FlowHandler) NewFlow(w http.ResponseWriter, r *http.Request) error {
    var f flow.Flow
    if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
        return err
    }

    uuid, err := h.manager.AddFlow(&f); 
    if err != nil {
        return err
    }

    dataResponse(w, h.parseFlow(&f, uuid))
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

    dataResponse(w, h.parseFlows(h.manager.GetFlows()))
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

    flow, _ := h.manager.GetFlow(*uuid)
    dataResponse(w, h.parseFlow(flow, uuid))
    return nil
}

func (h *FlowHandler) SetFlowAsMain(w http.ResponseWriter, r *http.Request) error {
    uuid, err := h.getUUIDFromBody(r)
    if err != nil {
        return err
    }
    h.manager.SetMain(*uuid)
    flow, _ := h.manager.GetFlow(*uuid)

    dataResponse(w, h.parseFlow(flow, uuid))
    return nil
}

func (s *Server) NewBroadcast(w http.ResponseWriter, r *http.Request) error {
    var body NewBroadcastPayload
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return err
	}
   
	comms, err := s.Store.GetAllCommunications(&body.Condition)
	if err != nil {
        return err
	}

    go s.flowHandler.manager.Broadcast(comms, body.Uuid)

	w.Header().Set("Content-Type", "application/json")
	res := struct {
		Success bool    `json:"success"`
		Count   int     `json:"count"`
	}{true, len(comms)}

	json.NewEncoder(w).Encode(res)
	return nil
}

func (h *FlowHandler) getUUIDFromBody(r *http.Request) (*uuid.UUID, error) {
    var body GetFlowPayload
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

    if _, err := h.manager.GetFlow(body.Uuid.UUID); err != nil{
        return nil, err
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
    if _, err := h.manager.GetFlow(uuid); err != nil{
        return nil, err
    }

    return &uuid, nil
}

func (h *FlowHandler) parseFlow(f *flow.Flow, uuid *uuid.UUID) *FlowResponse {
    return &FlowResponse{
        IsMain: h.manager.GetMain() == *uuid,
        Uuid: *uuid,
        Rules: *f,
    }
}

func (h *FlowHandler) parseFlows(flows map[uuid.UUID]flow.Flow) map[uuid.UUID]FlowResponse {
    res := make(map[uuid.UUID]FlowResponse)
    for uuid, rules := range flows {
        res[uuid] = *h.parseFlow(&rules, &uuid)
    }
    return res
}
