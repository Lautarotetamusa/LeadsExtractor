package handlers

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
	Uuid      uuid.UUID        `json:"uuid"`
	Condition store.QueryParam `json:"condition"`
}

type FlowHandler struct {
	manager   *flow.FlowManager
	commStore store.CommunicationStorer
}

type FlowResponse struct {
	Name   string      `json:"name"`
	IsMain bool        `json:"is_main"`
	Rules  []flow.Rule `json:"rules"`
	Uuid   uuid.UUID   `json:"uuid"`
}

func (h FlowHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/broadcast", HandleErrors(h.NewBroadcast)).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/mainFlow", HandleErrors(h.SetFlowAsMain)).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/actions", HandleErrors(h.GetConfig)).Methods(http.MethodGet, http.MethodOptions)

	r := router.PathPrefix("/flows").Subrouter()
	r.HandleFunc("", HandleErrors(h.NewFlow)).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/{uuid}", HandleErrors(h.UpdateFlow)).Methods(http.MethodPut, http.MethodOptions)
	r.HandleFunc("", HandleErrors(h.GetFlows)).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/main", HandleErrors(h.GetMainFlow)).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/{uuid}", HandleErrors(h.GetFlow)).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/{uuid}", HandleErrors(h.DeleteFlow)).Methods(http.MethodDelete, http.MethodOptions)
}

func NewFlowHandler(m *flow.FlowManager, comm store.CommunicationStorer) *FlowHandler {
	return &FlowHandler{
		manager: m,
        commStore: comm,
	}
}

func (h *FlowHandler) GetConfig(w http.ResponseWriter, r *http.Request) error {
	actions := h.manager.GetActions()
	dataResponse(w, actions)
	return nil
}

func (h *FlowHandler) GetFlows(w http.ResponseWriter, r *http.Request) error {
	flows := h.manager.GetAll()
	dataResponse(w, h.parseFlows(flows))
	return nil
}

func (h *FlowHandler) GetMainFlow(w http.ResponseWriter, r *http.Request) error {
	flow, err := h.manager.GetMain()
	if err != nil {
		return err
	}

	uuid := h.manager.GetMainUUID()
	dataResponse(w, h.parseFlow(flow, &uuid))
	return nil
}

func (h *FlowHandler) GetFlow(w http.ResponseWriter, r *http.Request) error {
	uuid, err := h.getUUIDFromParam(r)
	if err != nil {
		return err
	}
	flow, err := h.manager.GetOne(*uuid)
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

	uuid, err := h.manager.Add(&f)
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

	if *uuid == h.manager.GetMainUUID(){
	    return ErrBadRequest(fmt.Sprintf("no se puede eliminar el flow principal"))
	}

	if err := h.manager.Delete(*uuid); err != nil {
		return err
	}

	dataResponse(w, h.parseFlows(h.manager.GetAll()))
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

	if err := h.manager.Update(&f, *uuid); err != nil {
		return err
	}

	flow, _ := h.manager.GetOne(*uuid)
	dataResponse(w, h.parseFlow(flow, uuid))
	return nil
}

func (h *FlowHandler) SetFlowAsMain(w http.ResponseWriter, r *http.Request) error {
	uuid, err := h.getUUIDFromBody(r)
	if err != nil {
		return err
	}
	h.manager.SetMain(*uuid)
	flow, _ := h.manager.GetOne(*uuid)

	dataResponse(w, h.parseFlow(flow, uuid))
	return nil
}

func (s *FlowHandler) NewBroadcast(w http.ResponseWriter, r *http.Request) error {
	var body NewBroadcastPayload
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return jsonErr(err)
	}

	comms, err := s.commStore.GetDistinct(&body.Condition)
	if err != nil {
		return err
	}

	if _, err := s.manager.GetOne(body.Uuid); err != nil {
		return err
	}
	go s.manager.Broadcast(comms, body.Uuid)

	w.Header().Set("Content-Type", "application/json")
	res := struct {
		Success bool `json:"success"`
		Count   int  `json:"count"`
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

	if _, err := h.manager.GetOne(body.Uuid.UUID); err != nil {
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
	if _, err := h.manager.GetOne(uuid); err != nil {
		return nil, err
	}

	return &uuid, nil
}

func (h *FlowHandler) parseFlow(f *flow.Flow, uuid *uuid.UUID) *FlowResponse {
	return &FlowResponse{
		IsMain: h.manager.GetMainUUID() == *uuid,
		Uuid:   *uuid,
		Rules:  f.Rules,
		Name:   f.Name,
	}
}

func (h *FlowHandler) parseFlows(flows map[uuid.UUID]flow.Flow) map[uuid.UUID]FlowResponse {
	res := make(map[uuid.UUID]FlowResponse)
	for uuid, rules := range flows {
		res[uuid] = *h.parseFlow(&rules, &uuid)
	}
	return res
}
