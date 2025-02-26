package handlers

import (
	"encoding/json"
	"leadsextractor/store"
	"net/http"
	"strconv"

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

    r.Methods(http.MethodOptions)

	r.HandleFunc("", HandleErrors(h.GetAll)).Methods(http.MethodGet)
	r.HandleFunc("/{id}", HandleErrors(h.GetOne)).Methods(http.MethodGet)
	r.HandleFunc("", HandleErrors(h.Insert)).Methods(http.MethodPost)
	r.HandleFunc("/{id}", HandleErrors(h.Update)).Methods(http.MethodPut)

    // add an image to a property
	r.HandleFunc("/{id}/image", HandleErrors(h.AddImages)).Methods(http.MethodPost)
	r.HandleFunc("/{propId}/image/{imageId}", HandleErrors(h.DeleteImage)).Methods(http.MethodDelete, http.MethodOptions)
	// r.HandleFunc("/{id}/image", HandleErrors(h.GetImages)).Methods(http.MethodGet, http.MethodOptions)
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

    if err := h.storer.GetImages(prop); err != nil {
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

	if err = validate.Struct(prop); err != nil {
		return ErrBadRequest(err.Error())
	}

    for i, _ := range prop.Images {
        if err := validate.Var(prop.Images[i].Url, "required,url"); err != nil {
            return ErrBadRequest("'image must have a a valid url")
        }
    }

	if err := h.storer.Insert(&prop); err != nil {
		return err
	}

	createdResponse(w, "property created successfully", prop)
	return nil
}

func (h *PropertyHandler) AddImages(w http.ResponseWriter, r *http.Request) error {
    strId := mux.Vars(r)["id"]
    propId, err := strconv.ParseInt(strId, 10, 16)
    if err != nil {
        return ErrBadRequest("the id must be a integer")
    }

	var images []store.PropertyImage
	if err := json.NewDecoder(r.Body).Decode(&images); err != nil {
		return jsonErr(err)
	}

    for _, image := range images {
        if err := validate.Var(image.Url, "required,url"); err != nil {
            return ErrBadRequest("'image must have a a valid url")
        }
    } 

    if err := h.storer.InsertImages(propId, images); err != nil {
        return err
    }

    createdResponse(w, "images added successfully", images)
    return nil
}

func (h *PropertyHandler) DeleteImage(w http.ResponseWriter, r *http.Request) error {
    strId := mux.Vars(r)["propId"]
    propId, err := strconv.ParseInt(strId, 10, 16)
    if err != nil {
        return ErrBadRequest("the id must be a integer")
    }

    strId = mux.Vars(r)["imageId"]
    imageId, err := strconv.ParseInt(strId, 10, 16)
    if err != nil {
        return ErrBadRequest("the image id must be a integer")
    }

    return h.storer.DeleteImage(propId, imageId)
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

	if err := validate.Struct(prop); err != nil {
		return ErrBadRequest(err.Error())
	}

	if err := h.storer.Update(&prop); err != nil {
		return err
	}

	createdResponse(w, "property updated successfully", prop)
	return nil
}
