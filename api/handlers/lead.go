package handlers

import (
	"encoding/json"
	"leadsextractor/models"
	"leadsextractor/pkg/numbers"
	"leadsextractor/store"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type LeadHandler struct {
	storer store.LeadStorer
}

func NewLeadHandler(s store.LeadStorer) *LeadHandler {
	return &LeadHandler{
		storer: s,
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
	leads, err := h.storer.GetAll()
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

	lead, err := h.storer.GetOne(*phone)
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

	validate := validator.New()
	if err = validate.Struct(createLead); err != nil {
		return ErrBadRequest(err.Error())
	}

	lead, err := h.storer.Insert(&createLead)
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

	lead, err := h.storer.GetOne(*phone)
	if err != nil {
		return err
	}

	var updateLead models.UpdateLead
	if err := json.NewDecoder(r.Body).Decode(&updateLead); err != nil {
		return jsonErr(err)
	}

	updateLeadsFields(lead, updateLead)
	validate := validator.New()
	if err := validate.Struct(lead); err != nil {
		return ErrBadRequest(err.Error())
	}

	if err := h.storer.Update(lead); err != nil {
		return err
	}

	createdResponse(w, "Lead actualizado correctamente", lead)
	return nil
}

func updateLeadsFields(lead *models.Lead, update models.UpdateLead) {
	if update.Name != "" {
		lead.Name = update.Name
	}
	if update.Email.Valid {
		lead.Email = update.Email
	}
	if update.Cotizacion != "" {
		lead.Cotizacion = update.Cotizacion
	}
}
