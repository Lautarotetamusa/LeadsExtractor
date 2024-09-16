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

type FlowHandler struct {
    manager *flow.FlowManager
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
    dataResponse(w, flows)
    return nil
}

func (h *FlowHandler) GetMainFlow(w http.ResponseWriter, r *http.Request) error {
    flow, err := h.manager.GetMainFlow()
    if err != nil {
        return err
    }

    dataResponse(w, flow)
    return nil
}

func (h *FlowHandler) GetFlow(w http.ResponseWriter, r *http.Request) error {
    flow, err := h.getFlowFromParam(r)
    if err != nil {
        return err
    }

    dataResponse(w, flow)
    return nil
}

func (h *FlowHandler) NewFlow(w http.ResponseWriter, r *http.Request) error {
    var f flow.Flow
    if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
        return err
    }

    err := h.manager.AddFlow(&f); 
    if err != nil {
        return err
    }

    dataResponse(w, f)
    return nil
}

func (h *FlowHandler) DeleteFlow(w http.ResponseWriter, r *http.Request) error {
    flow, err := h.getFlowFromParam(r)
    if err != nil {
        return err
    }
    if err := h.manager.DeleteFlow(flow.Uuid); err != nil {
        return err
    }

    dataResponse(w, h.manager.GetFlows())
    return nil
}

func (h *FlowHandler) UpdateFlow(w http.ResponseWriter, r *http.Request) error {
    var f flow.Flow
    if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
        return err
    }

    flow, err := h.getFlowFromParam(r)
    if err != nil {
        return err
    }

    if err := h.manager.UpdateFlow(&f, flow.Uuid); err != nil {
        return err
    }

    dataResponse(w, f)
    return nil
}

func (h *FlowHandler) SetFlowAsMain(w http.ResponseWriter, r *http.Request) error {
    flow, err := h.getFlowFromBody(r)
    if err != nil {
        return err
    }
    h.manager.SetMain(flow.Uuid)

    dataResponse(w, flow)
    return nil
}

func (s *Server) NewBroadcast(w http.ResponseWriter, r *http.Request) error {
    var body struct {
        Uuid        uuid.UUID           `json:"uuid"`
        Condition   store.QueryParam    `json:"condition"`
    }

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

func (h *FlowHandler) getFlowFromBody(r *http.Request) (*flow.Flow, error) {
    var body struct {
        Uuid uuid.NullUUID `json:"uuid"`
    }

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

    flow, err := h.manager.GetFlow(body.Uuid.UUID); 
    if err != nil{
        return nil, err
    }

    return flow, nil
}

func (h *FlowHandler) getFlowFromParam(r *http.Request) (*flow.Flow, error) {
    uuidStr, exists := mux.Vars(r)["uuid"]
    if !exists {
        return nil, fmt.Errorf("se necesita pasar un uuid")
    }
    uuid, err := uuid.Parse(uuidStr)
    if err != nil {
        return nil, err
    }
    flow, err := h.manager.GetFlow(uuid); 
    if err != nil{
        return nil, err
    }

    return flow, nil
}
