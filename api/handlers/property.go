package handlers

import (
	"encoding/json"
	"leadsextractor/store"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type PropertyHandler struct {
    storer store.PropertyPortalStore
}

func NewPropertyHandler(s store.PropertyPortalStore) *PropertyHandler {
	return &PropertyHandler{
		storer: s,
	}
}

func (h *PropertyHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/property").Subrouter()

	r.HandleFunc("", HandleErrors(h.GetAll)).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/{id}", HandleErrors(h.GetOne)).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("", HandleErrors(h.Insert)).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/{id}", HandleErrors(h.Update)).Methods(http.MethodPut, http.MethodOptions)
}

func (h *PropertyHandler) GetAll(w http.ResponseWriter, r *http.Request) error {
	props, err := h.storer.GetAll()
	if err != nil {
		return err
	}

	dataResponse(w, props)
	return nil
}

func (h *PropertyHandler) GetOne(w http.ResponseWriter, r *http.Request) error {
    strId := mux.Vars(r)["id"]
    id, err := strconv.ParseInt(strId, 10, 16)
    if err != nil {
        return ErrBadRequest("the id must be a integer")
    }

	prop, err := h.storer.GetOne(id)
	if err != nil {
		return err
	}

	dataResponse(w, prop)
	return nil
}

func (h *PropertyHandler) Insert(w http.ResponseWriter, r *http.Request) error {
	var prop store.PortalProp
	err := json.NewDecoder(r.Body).Decode(&prop)
	if err != nil {
		return jsonErr(err)
	}

	validate := validator.New()
	if err = validate.Struct(prop); err != nil {
		return ErrBadRequest(err.Error())
	}

	if err := h.storer.Insert(&prop); err != nil {
		return err
	}

	createdResponse(w, "property created successfully", prop)
	return nil
}

func (h *PropertyHandler) Update(w http.ResponseWriter, r *http.Request) error {
    strId := mux.Vars(r)["id"]
    id, err := strconv.ParseInt(strId, 10, 16)
    if err != nil {
        return ErrBadRequest("the id must be a integer")
    }

	var prop store.PortalProp
	if err := json.NewDecoder(r.Body).Decode(&prop); err != nil {
		return jsonErr(err)
	}
    prop.ID = id

	validate := validator.New()
	if err := validate.Struct(prop); err != nil {
		return ErrBadRequest(err.Error())
	}

	if err := h.storer.Update(&prop); err != nil {
		return err
	}

	createdResponse(w, "property updated successfully", prop)
	return nil
}
