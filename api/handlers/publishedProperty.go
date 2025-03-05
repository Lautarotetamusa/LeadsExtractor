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
	storer store.PublishedPropertyStorer
}

func NewPublishedPropertyHandler(store store.PublishedPropertyStorer) *PublishedPropertyHandler {
	return &PublishedPropertyHandler{storer: store}
}

func (h *PublishedPropertyHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/property/{propId}/publications").Subrouter()
    r.Methods(http.MethodOptions)

    // Property publications
	router.HandleFunc("/publish", HandleErrors(h.Publish)).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("", HandleErrors(h.GetPublications)).Methods(http.MethodGet)
	r.HandleFunc("/{portal}", HandleErrors(h.GetPublication)).Methods(http.MethodGet)
	r.HandleFunc("/{portal}/status", HandleErrors(h.UpdateStatus)).Methods(http.MethodPut)
}

// CreateHandler handles property creation
func (h *PublishedPropertyHandler) Publish(w http.ResponseWriter, r *http.Request) error {
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

    prop, err := h.storer.GetOne(pp.Portal.String, int64(pp.PropertyID.Int16))
    if prop != nil {
        // If the property its already publicated on this portal
        // and the property status its not "completed" return a error
        if prop.Status != store.StatusCompleted {
            return ErrBadRequest("the property has a publication in progress, wait until the end")
        }

        // The property exists but publishing completed, then republish
        err := h.storer.UpdateStatus(pp.Portal.String, int64(pp.PropertyID.Int16), store.StatusInProgress)
        if err != nil {
            return err
        }
    } else {
        // If the property does not have a publication on the portal, create it
        if err := h.storer.Create(&pp); err != nil {
            return err
        }
    }

    // Get the updated_at and created_at fields
    prop, err = h.storer.GetOne(pp.Portal.String, int64(pp.PropertyID.Int16))
    if err != nil {
        return err
    }
    createdResponse(w, "property publishing process has started", prop)
    return nil
}

func (h *PublishedPropertyHandler) GetPublications(w http.ResponseWriter, r *http.Request) error {
    strId := mux.Vars(r)["propId"]
    id, err := strconv.ParseInt(strId, 10, 16)
    if err != nil {
        return InvalidPropID
    }

    props, err := h.storer.GetAllByProp(id)
    if err != nil {
        return err
    }

    dataResponse(w, props)
    return nil
}

func (h *PublishedPropertyHandler) GetPublication(w http.ResponseWriter, r *http.Request) error {
	portal := mux.Vars(r)["portal"]
	propertyIDStr := mux.Vars(r)["propId"]
	
	propId, err := strconv.Atoi(propertyIDStr)
	if err != nil {
		return ErrBadRequest("Invalid property ID")
	}

	pp, err := h.storer.GetOne(portal, int64(propId))
	if err != nil {
		return err
	}

    dataResponse(w, pp)
    return nil
}

func (h *PublishedPropertyHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) error {
	portal := mux.Vars(r)["portal"]
	propertyIDStr := mux.Vars(r)["propId"]
	
	propId, err := strconv.Atoi(propertyIDStr)
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

	if err := h.storer.UpdateStatus(portal, int64(propId), statusUpdate.Status); err != nil {
		return err
	}

    messageResponse(w, "status updated successfully")
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
