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

var validChannels = []string{"ivr", "whatsapp", "inbox"}
func validateChannel(utm *models.UtmDefinition) error {
    if !slices.Contains(validChannels, utm.Channel.String) {
        return fmt.Errorf("el channel no es valido. validos: %s", strings.Join(validChannels, ", "))
    }
    return nil
}

func (h UTMHandler) GetAll(w http.ResponseWriter, r *http.Request) error {
    var utms []models.UtmDefinition
	if err := h.storer.GetAll(&utms); err != nil {
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
        return fmt.Errorf("no se encontro el utm con id %d", id)
	}

	dataResponse(w, utm)
	return nil
}

func (h UTMHandler) Insert(w http.ResponseWriter, r *http.Request) error {
	var utm models.UtmDefinition
    defer r.Body.Close()
    if err := json.NewDecoder(r.Body).Decode(&utm); err != nil {
        return err
    }

    if err := validateChannel(&utm); err != nil {
        return err
    }

    isValid := true
    for _, r := range utm.Code {
        if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
            isValid = false
        }
    }

    if !isValid {
        return fmt.Errorf("el codigo solo puede contener caracteres alphanumericos")
    }
    utm.Code = strings.ToUpper(utm.Code)

    id, err := h.storer.Insert(&utm); 
    if err != nil {
		return err
	}
    utm.Id = int(id)

	successResponse(w, "Utm creado correctamente", utm)
	return nil
}

func (h UTMHandler) Update(w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["id"]
    id, err := strconv.Atoi(idStr)
    utm, err := h.storer.GetOne(id)
    if err != nil {
        return fmt.Errorf("no existe utm con id %d", id)
    }

    defer r.Body.Close()
    if err := json.NewDecoder(r.Body).Decode(&utm); err != nil {
		return err
    }

    if err := validateChannel(utm); err != nil {
        return err
    }

    if err := h.storer.Update(utm); err != nil {
		return err
	}

	successResponse(w, "Utm actualizado correctamente", utm)
	return nil
}
