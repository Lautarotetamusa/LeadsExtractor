package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"leadsextractor/models"
	"leadsextractor/service"

	"github.com/gorilla/mux"
)

type UTMHandler struct {
	service *service.UTMService
}

func NewUTMHandler(s *service.UTMService) *UTMHandler {
	return &UTMHandler{
		service: s,
	}
}

func (h UTMHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/utm").Subrouter()

	r.HandleFunc("", HandleErrors(h.GetAll)).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/{id}", HandleErrors(h.GetOne)).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("", HandleErrors(h.Insert)).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/{id}", HandleErrors(h.Update)).Methods(http.MethodPut, http.MethodOptions)
	r.HandleFunc("/{id}", HandleErrors(h.Delete)).Methods(http.MethodDelete, http.MethodOptions)
}

func (h UTMHandler) GetAll(w http.ResponseWriter, r *http.Request) error {
	utms, err := h.service.GetAll()
	if err != nil {
		return err
	}

	dataResponse(w, utms)
	return nil
}

func (h UTMHandler) GetOne(w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	utm, err := h.service.GetOne(id)
	if err != nil {
		return err
	}

	dataResponse(w, utm)
	return nil
}

func (h UTMHandler) Insert(w http.ResponseWriter, r *http.Request) error {
	var utm models.UtmDefinition
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&utm); err != nil {
		return jsonErr(err)
	}

	if err := h.service.Insert(&utm); err != nil {
		return err
	}

	createdResponse(w, "utm created successfully", utm)
	return nil
}

func (h UTMHandler) Update(w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	utm, err := h.service.GetOne(id)
	if err != nil {
		return err
	}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&utm); err != nil {
		return jsonErr(err)
	}

	if err := h.service.Update(utm); err != nil {
		return err
	}

	createdResponse(w, "utm updated successfully", utm)
	return nil
}

func (h UTMHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	if err := h.service.Delete(id); err != nil {
		return err
	}

	messageResponse(w, "utm deleted successfully")
	return nil
}
