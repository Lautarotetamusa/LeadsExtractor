package handlers

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"leadsextractor/models"
	"leadsextractor/store"
	"net/http"
	"strconv"
	"unicode"

	"github.com/gorilla/mux"
)

type UTMHandler struct {
    storer  store.UTMStorer
}

var validChannels = []string{"ivr", "whatsapp", "inbox"}

func NewUTMHandler(s store.UTMStorer) *UTMHandler {
    return &UTMHandler{
        storer: s,
    }
}

func (h UTMHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/utm", HandleErrors(h.GetAll)).Methods("GET", "OPTIONS")
	router.HandleFunc("/utm/{id}", HandleErrors(h.GetOne)).Methods("GET", "OPTIONS")
	router.HandleFunc("/utm", HandleErrors(h.Insert)).Methods("POST", "OPTIONS")
	router.HandleFunc("/utm/{id}", HandleErrors(h.Update)).Methods("PUT", "OPTIONS")
}

func validateChannel(utm *models.UtmDefinition) error {
    if !slices.Contains(validChannels, utm.Channel.String) {
        return fmt.Errorf("the channel must be one of %s", strings.Join(validChannels, ", "))
    }
    return nil
}

func (h UTMHandler) GetAll(w http.ResponseWriter, r *http.Request) error {
    utms, err := h.storer.GetAll();
	if err != nil {
		return err
	}

	dataResponse(w, utms)
	return nil
}

func (h UTMHandler) GetOne(w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["id"]
    id, err := strconv.Atoi(idStr)

    utm, err := h.storer.GetOne(id)
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
        return ErrBadRequest(err.Error())
    }

    if err := validateChannel(&utm); err != nil {
        return ErrBadRequest(err.Error())
    }

    if utm.Code == "" {
        return ErrBadRequest("code is required")
    }

    isValid := true
    for _, r := range utm.Code {
        if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
            isValid = false
        }
    }

    if !isValid {
        return ErrBadRequest("code can only contains alphanumeric characters")
    }
    utm.Code = strings.ToUpper(utm.Code)

    id, err := h.storer.Insert(&utm); 
    if err != nil {
		return err
	}
    utm.Id = int(id)

	createdResponse(w, "utm created successfully", utm)
	return nil
}

func (h UTMHandler) Update(w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["id"]
    id, err := strconv.Atoi(idStr)
    utm, err := h.storer.GetOne(id)
    if err != nil {
        return err
    }

    defer r.Body.Close()
    if err := json.NewDecoder(r.Body).Decode(&utm); err != nil {
		return ErrBadRequest(err.Error())
    }

    if err := validateChannel(utm); err != nil {
		return ErrBadRequest(err.Error())
    }

    if err := h.storer.Update(utm); err != nil {
		return err
	}

	createdResponse(w, "utm updated successfully", utm)
	return nil
}
