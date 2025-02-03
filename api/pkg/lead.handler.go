package pkg

import (
	"encoding/json"
	"fmt"
	"leadsextractor/models"
	"leadsextractor/numbers"
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
	router.HandleFunc("/lead", HandleErrors(h.GetAll)).Methods("GET", "OPTIONS")
	router.HandleFunc("/lead/{phone}", HandleErrors(h.GetOne)).Methods("GET", "OPTIONS")
	router.HandleFunc("/lead", HandleErrors(h.Insert)).Methods("POST", "OPTIONS")
	router.HandleFunc("/lead/{phone}", HandleErrors(h.Update)).Methods("PUT", "OPTIONS")
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
        return fmt.Errorf("el numero %s no es un telefono valido", phone)
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
    fmt.Println("HOLA")
	err := json.NewDecoder(r.Body).Decode(&createLead)
	if err != nil {
        fmt.Printf("%#v\n", err)
		return err
	}

	validate := validator.New()
	if err = validate.Struct(createLead); err != nil {
		return err
	}

	lead, err := h.storer.Insert(&createLead)
	if err != nil {
		return err
	}

    fmt.Printf("%v\n", lead)

    w.WriteHeader(http.StatusCreated)
	successResponse(w, "Lead creado correctamente", lead)
	return nil
}

func (h *LeadHandler) Update(w http.ResponseWriter, r *http.Request) error {
	phone, err := numbers.NewPhoneNumber(mux.Vars(r)["phone"])
    if err != nil {
        return fmt.Errorf("el numero %s no es un telefono valido", phone.String())
    }

	var updateLead models.UpdateLead
	if err := json.NewDecoder(r.Body).Decode(&updateLead); err != nil {
		return err
	}

	lead := models.Lead{
		Phone: *phone,
		Name:  updateLead.Name,
		Email: updateLead.Email,
	}

	validate := validator.New()
    if err := validate.Struct(lead); err != nil {
		return err
	}

    if err := h.storer.Update(&lead, *phone); err != nil {
		return err
	}

	successResponse(w, "Lead actualizado correctamente", lead)
	return nil
}
