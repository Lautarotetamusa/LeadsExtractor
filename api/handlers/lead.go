package handlers

import (
	"database/sql"
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
    service *LeadService
}

type LeadService struct {
    storer store.LeadStorer
}

func NewLeadService(s store.LeadStorer) *LeadService {
    return &LeadService{
        storer: s,
    }
}

func NewLeadHandler(s *LeadService) *LeadHandler {
    return &LeadHandler{
        service: s,
    }
}

func (h LeadHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/lead", HandleErrors(h.GetAll)).Methods("GET", "OPTIONS")
	router.HandleFunc("/lead/{phone}", HandleErrors(h.GetOne)).Methods("GET", "OPTIONS")
	router.HandleFunc("/lead", HandleErrors(h.Insert)).Methods("POST", "OPTIONS")
	router.HandleFunc("/lead/{phone}", HandleErrors(h.Update)).Methods("PUT", "OPTIONS")
}

// GetOrInsert get the lead with phone c.Telefono in case that exists, otherwise creates one
func (s *LeadService) GetOrInsert(rr *store.RoundRobin, c *models.Communication) (*models.Lead, error) {
	lead, err := s.storer.GetOne(c.Telefono)

    // The lead does not exists
	if err == sql.ErrNoRows {
		c.IsNew = true
		c.Asesor = rr.Next()

		lead, err = s.storer.Insert(&models.CreateLead{
			Name:        c.Nombre,
			Phone:       c.Telefono,
			Email:       c.Email,
			AsesorPhone: c.Asesor.Phone,
            Cotizacion:  c.Cotizacion,
		})

		if err != nil {
			return nil, err
		}
	} else if err != nil {
        if err := s.storer.Update(lead, lead.Phone); err != nil {
            return nil, err
        }

		return nil, err
	}

	return lead, nil
}

func (h *LeadHandler) GetAll(w http.ResponseWriter, r *http.Request) error {
	leads, err := h.service.storer.GetAll()
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

	lead, err := h.service.storer.GetOne(*phone)
	if err != nil {
        if err == sql.ErrNoRows {
            return ErrNotFound(fmt.Sprintf("the lead with phone %s does not exists", phone))
        }
		return err
	}

	dataResponse(w, lead)
	return nil
}

func (h *LeadHandler) Insert(w http.ResponseWriter, r *http.Request) error {
	var createLead models.CreateLead
	err := json.NewDecoder(r.Body).Decode(&createLead)
	if err != nil {
		return ErrBadRequest(err.Error())
	}

	validate := validator.New()
	if err = validate.Struct(createLead); err != nil {
		return ErrBadRequest(err.Error())
	}

	lead, err := h.service.storer.Insert(&createLead)
	if err != nil {
		return err
	}

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

    if err := h.service.storer.Update(&lead, *phone); err != nil {
		return err
	}

	successResponse(w, "Lead actualizado correctamente", lead)
	return nil
}
