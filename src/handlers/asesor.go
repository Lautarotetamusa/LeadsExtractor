package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"leadsextractor/models"
	"leadsextractor/service"

	"github.com/gorilla/mux"
)

type AsesorHandler struct {
	service *service.AsesorService
}

func NewAsesorHandler(s *service.AsesorService) *AsesorHandler {
	return &AsesorHandler{
		service: s,
	}
}

func (h *AsesorHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/asesor").Subrouter()

	r.HandleFunc("", HandleErrors(h.GetAll)).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/{phone}", HandleErrors(h.GetOne)).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("", HandleErrors(h.Insert)).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/{phone}", HandleErrors(h.Update)).Methods(http.MethodPut, http.MethodOptions)
	r.HandleFunc("/{phone}", HandleErrors(h.Delete)).Methods(http.MethodDelete, http.MethodOptions)
	r.HandleFunc("/{phone}/reasign", HandleErrors(h.Reasign)).Methods(http.MethodPut, http.MethodOptions)
	r.HandleFunc("/{phone}/leads", HandleErrors(h.GetLeads)).Methods(http.MethodGet, http.MethodOptions)
}

func (h *AsesorHandler) GetAll(w http.ResponseWriter, r *http.Request) error {
	asesores, err := h.service.GetAll()
	if err != nil {
		return err
	}

	dataResponse(w, asesores)
	return nil
}

func (h *AsesorHandler) GetOne(w http.ResponseWriter, r *http.Request) error {
	phone := mux.Vars(r)["phone"]

	asesor, err := h.service.GetOne(phone)
	if err != nil {
		return err
	}

	dataResponse(w, asesor)
	return nil
}

func (h *AsesorHandler) GetLeads(w http.ResponseWriter, r *http.Request) error {
	phone := mux.Vars(r)["phone"]

	leads, err := h.service.GetLeads(phone)
	if err != nil {
		return err
	}

	dataResponse(w, leads)
	return nil
}

func (h *AsesorHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	phone := mux.Vars(r)["phone"]

	if err := h.service.Delete(phone); err != nil {
		return err
	}

	messageResponse(w, "Asesor eliminado con exito")
	return nil
}

func (h *AsesorHandler) Insert(w http.ResponseWriter, r *http.Request) error {
	var asesor models.Asesor
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&asesor); err != nil {
		return jsonErr(err)
	}

	if err := h.service.Insert(&asesor); err != nil {
		return err
	}

	createdResponse(w, "Asesor creado correctamente", asesor)
	return nil
}

func (h *AsesorHandler) Reasign(w http.ResponseWriter, r *http.Request) error {
	phone := mux.Vars(r)["phone"]

	n, err := h.service.ReasignLeads(phone)
	if err != nil {
		return err
	}

	messageResponse(w, fmt.Sprintf("se reasignaron un total de %d leads", n))
	return nil
}

func (h *AsesorHandler) Update(w http.ResponseWriter, r *http.Request) error {
	phone := mux.Vars(r)["phone"]

	var updateAsesor models.UpdateAsesor
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&updateAsesor); err != nil {
		return jsonErr(err)
	}

	asesor, err := h.service.Update(phone, updateAsesor)
	if err != nil {
		return err
	}

	createdResponse(w, "Asesor actualizado correctamente", asesor)
	return nil
}
