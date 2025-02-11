package handlers

import (
	"database/sql"
	"encoding/json"
	"leadsextractor/models"
	"leadsextractor/pkg/numbers"
	"leadsextractor/pkg/roundrobin"
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
    r := router.PathPrefix("/lead").Subrouter()

	r.HandleFunc("", HandleErrors(h.GetAll)).Methods(http.MethodGet)
	r.HandleFunc("/{phone}", HandleErrors(h.GetOne)).Methods(http.MethodGet)
	r.HandleFunc("", HandleErrors(h.Insert)).Methods(http.MethodPost)
	r.HandleFunc("/{phone}", HandleErrors(h.Update)).Methods(http.MethodPut)
}

// GetOrInsert get the lead with phone c.Telefono in case that exists, otherwise creates one
func (s *LeadService) GetOrInsert(rr *roundrobin.RoundRobin[models.Asesor], c *models.Communication) (*models.Lead, error) {
	lead, err := s.storer.GetOne(c.Telefono)

    // The lead does not exists
	if err == sql.ErrNoRows {
		c.IsNew = true
		c.Asesor = *rr.Next()

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
        if err := s.storer.Update(lead); err != nil {
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
        return ErrBadRequest(err.Error())
    }

	lead, err := h.service.storer.GetOne(*phone)
	if err != nil {
        return err
	}

	dataResponse(w, lead)
	return nil
}

func (h *LeadHandler) Insert(w http.ResponseWriter, r *http.Request) error {
	var createLead models.CreateLead
	err := json.NewDecoder(r.Body).Decode(&createLead)
	if err != nil {
        return jsonErr(err)
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
	createdResponse(w, "Lead creado correctamente", lead)
	return nil
}

func (h *LeadHandler) Update(w http.ResponseWriter, r *http.Request) error {
	phone, err := numbers.NewPhoneNumber(mux.Vars(r)["phone"])
    if err != nil {
        return ErrBadRequest(err.Error())
    }

    lead, err := h.service.storer.GetOne(*phone)
    if err != nil {
        return err
    }

	var updateLead models.UpdateLead
	if err := json.NewDecoder(r.Body).Decode(&updateLead); err != nil {
        return jsonErr(err)
	}

    updateLeadsFields(lead, updateLead)
	validate := validator.New()
	   if err := validate.Struct(lead); err != nil {
		return ErrBadRequest(err.Error())
	}

    if err := h.service.storer.Update(lead); err != nil {
		return err
	}

	createdResponse(w, "Lead actualizado correctamente", lead)
	return nil
}

func updateLeadsFields(lead *models.Lead, update models.UpdateLead) {
	if update.Name != "" {
		lead.Name = update.Name
	}
	if update.Email.Valid {
		lead.Email = update.Email
	}
	if update.Cotizacion != "" {
		lead.Cotizacion = update.Cotizacion
	}
}
