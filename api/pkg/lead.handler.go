package pkg

import (
	"encoding/json"
	"fmt"
	"leadsextractor/models"
	"leadsextractor/numbers"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

func (s *Server) GetAll(w http.ResponseWriter, r *http.Request) error {
	leads, err := s.Store.GetAll()
	if err != nil {
		return err
	}

	dataResponse(w, leads)
	return nil
}

func (s *Server) GetOne(w http.ResponseWriter, r *http.Request) error {
	phone, err := numbers.NewPhoneNumber(mux.Vars(r)["phone"])
    if err != nil {
        return fmt.Errorf("el numero %s no es un telefono valido", phone)
    }

	lead, err := s.Store.GetOne(*phone)
	if err != nil {
		return err
	}

	dataResponse(w, lead)
	return nil
}

func (s *Server) Insert(w http.ResponseWriter, r *http.Request) error {
	var createLead models.CreateLead
	err := json.NewDecoder(r.Body).Decode(&createLead)
	if err != nil {
		return err
	}

	validate := validator.New()
	if err = validate.Struct(createLead); err != nil {
		return err
	}

	lead, err := s.Store.Insert(&createLead)
	if err != nil {
		return err
	}

	successResponse(w, "Lead creado correctamente", lead)
	return nil
}

func (s *Server) Update(w http.ResponseWriter, r *http.Request) error {
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

    if err := s.Store.Update(&lead, *phone); err != nil {
		return err
	}

	successResponse(w, "Lead actualizado correctamente", lead)
	return nil
}
