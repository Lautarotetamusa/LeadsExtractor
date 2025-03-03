package handlers

import (
	"encoding/json"
	"leadsextractor/store"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type PublishPayload struct {
    Properties []store.PublishedProperty `json:"properties" validate:"required"`
}

type PublishedPropertyHandler struct {
	store       store.PublishedPropertyStorer
    propStore   store.PropertyPortalStore
}

func NewPublishedPropertyHandler(store store.PublishedPropertyStorer) *PublishedPropertyHandler {
	return &PublishedPropertyHandler{store: store}
}

func (h *PublishedPropertyHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/publishedProperty").Subrouter()

    r.Methods(http.MethodOptions)
    r.HandleFunc("", HandleErrors(h.Create)).Methods(http.MethodPost)
    // r.HandleFunc("", HandleErrors(h.GetAll)).Methods(http.MethodGet)
	r.HandleFunc("/{portal}/{propertyID}", HandleErrors(h.GetOne)).Methods(http.MethodGet)
	r.HandleFunc("/{portal}/{propertyID}/status", HandleErrors(h.UpdateStatus)).Methods(http.MethodPut)
}

// CreateHandler handles property creation
func (h *PublishedPropertyHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var payload PublishPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return jsonErr(err)
	}

    if err := validate.Struct(payload); err != nil {
        return ErrBadRequest(err.Error())
    }

    if len(payload.Properties) == 0 {
        return ErrBadRequest("must publish at least one property")
    }

    pp := store.PublishedProperty{
        PropertyID: payload.Properties[0].PropertyID,
        Portal: payload.Properties[0].Portal,
        Status: store.StatusInProgress,
    }

	if err := h.store.Create(&pp); err != nil {
		return err
	}

    createdResponse(w, "property publishing process has started", pp)
    return nil
}

func (h *PublishedPropertyHandler) GetOne(w http.ResponseWriter, r *http.Request) error {
	portal := r.PathValue("portal")
	propertyIDStr := r.PathValue("propertyID")
	
	propertyID, err := strconv.Atoi(propertyIDStr)
	if err != nil {
		return ErrBadRequest("Invalid property ID")
	}

	pp, err := h.store.GetOne(portal, propertyID)
	if err != nil {
		return err
	}

    dataResponse(w, pp)
    return nil
}

func (h *PublishedPropertyHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) error {
	portal := r.PathValue("portal")
	propertyIDStr := r.PathValue("propertyID")
	
	propertyID, err := strconv.Atoi(propertyIDStr)
	if err != nil {
		return ErrBadRequest("Invalid property ID")
	}

	var statusUpdate struct {
		Status store.PublishedStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&statusUpdate); err != nil {
		return jsonErr(err)
	}

	if !validStatus(statusUpdate.Status) {
		return ErrBadRequest("Invalid status value")
	}

	if err := h.store.UpdateStatus(portal, propertyID, statusUpdate.Status); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
    return nil
}

func validStatus(s store.PublishedStatus) bool {
	switch s {
	case store.StatusInProgress, store.StatusCompleted, store.StatusFailed:
		return true
	default:
		return false
	}
}
