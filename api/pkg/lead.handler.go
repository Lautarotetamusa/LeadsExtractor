package pkg

import (
	"encoding/json"
	"io/ioutil"
	"leadsextractor/models"
	"leadsextractor/store"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type LeadHandler struct{
    Store store.LeadStorer
}

func (h *LeadHandler) GetAll(w http.ResponseWriter, r *http.Request) error{
    leades, err := h.Store.GetAll()
    if err != nil {
        return err
    }

    dataResponse(w, r, leades)
    return nil
}

func (h *LeadHandler) GetOne(w http.ResponseWriter, r *http.Request) error{
    phone := mux.Vars(r)["phone"]

    lead, err := h.Store.GetOne(phone)
    if err != nil {
        return err
    }

    dataResponse(w, r, lead)
    return nil
}

func (h *LeadHandler) Insert(w http.ResponseWriter, r *http.Request) error{
    var createLead models.CreateLead
    reqBody, err := ioutil.ReadAll(r.Body)
    if err != nil {
        return err
    }

    if err = json.Unmarshal(reqBody, &createLead); err != nil{
        return err
    }

    validate := validator.New()
    if err = validate.Struct(createLead); err != nil{
        return err
    }

    lead, err := h.Store.Insert(&createLead)
    if err != nil{
        return err
    }

    successResponse(w, r, "Lead creado correctamente", lead)
    return nil
}

func (h *LeadHandler) Update(w http.ResponseWriter, r *http.Request) error{
    phone := mux.Vars(r)["phone"]

    var updateLead models.UpdateLead
    reqBody, err := ioutil.ReadAll(r.Body)
    if err != nil {
        return err
    }

    if err = json.Unmarshal(reqBody, &updateLead); err != nil{
        return err
    }
    lead := models.Lead{
        Phone: phone,
        Name: updateLead.Name,
        Email: updateLead.Email,
    }

    validate := validator.New()
    if err = validate.Struct(lead); err != nil{
        return err
    }

    _, err = h.Store.Update(&lead, phone)
    if err != nil{
        return err
    }

    successResponse(w, r, "Lead actualizado correctamente", lead)
    return nil
}
