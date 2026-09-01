package handlers

import (
	"encoding/json"
	"net/http"

	"leadsextractor/models"
	"leadsextractor/pkg/numbers"
	"leadsextractor/service"

	"github.com/gorilla/mux"
)

type LeadHandler struct {
	service *service.LeadService
}

func NewLeadHandler(s *service.LeadService) *LeadHandler {
	return &LeadHandler{
		service: s,
	}
}

func (h LeadHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/lead").Subrouter()

	r.HandleFunc("", HandleErrors(h.GetAll)).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/{phone}", HandleErrors(h.GetOne)).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("", HandleErrors(h.Insert)).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/{phone}", HandleErrors(h.Update)).Methods(http.MethodPut, http.MethodOptions)
}

func (h *LeadHandler) GetAll(w http.ResponseWriter, r *http.Request) error {
	leads, err := h.service.GetAll()
	if err != nil {
		return err
	}

	dataResponse(w, leads)
	return nil
}

func (h *LeadHandler) GetOne(w http.ResponseWriter, r *http.Request) error {
	phone, err := numbers.NewPhoneNumber(mux.Vars(r)["phone"])
	if err != nil {
		return ErrBadRequest(err.Error())
	}

	lead, err := h.service.GetOne(*phone)
	if err != nil {
		return err
	}

	dataResponse(w, lead)
	return nil
}

func (h *LeadHandler) Insert(w http.ResponseWriter, r *http.Request) error {
	var createLead models.CreateLead
	err := json.NewDecoder(r.Body).Decode(&createLead)
	if err != nil {
		return jsonErr(err)
	}

	lead, err := h.service.Insert(&createLead)
	if err != nil {
		return err
	}

	createdResponse(w, "Lead creado correctamente", lead)
	return nil
}

func (h *LeadHandler) Update(w http.ResponseWriter, r *http.Request) error {
	phone, err := numbers.NewPhoneNumber(mux.Vars(r)["phone"])
	if err != nil {
		return ErrBadRequest(err.Error())
	}

	var updateLead models.UpdateLead
	if err := json.NewDecoder(r.Body).Decode(&updateLead); err != nil {
		return jsonErr(err)
	}

	lead, err := h.service.Update(*phone, updateLead)
	if err != nil {
		return err
	}

	createdResponse(w, "Lead actualizado correctamente", lead)
	return nil
}
