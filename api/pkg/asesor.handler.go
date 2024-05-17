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

type ReasignData struct {
    Phone   string `json:"phone"`
}

type AsesorHandler struct{
    Store store.AsesorStorer
}

func (s *Server) UpdateStatuses(w http.ResponseWriter, r *http.Request) error{
    var asesores []models.Asesor

    defer r.Body.Close()
    reqBody, err := ioutil.ReadAll(r.Body)
    if err != nil {
        return err
    }

    err = json.Unmarshal(reqBody, &asesores)
    if err != nil{
        return err
    }

    for i := range asesores{
        err = s.asesorHandler.Store.Update(&asesores[i], asesores[i].Phone)
        if err != nil{
            return err
        }
    }

    s.roundRobin.SetAsesores(s.db)

    successResponse(w, r, "Asesores actualizados correctamente", nil)
    return nil
}

/*
func (h *AsesorHandler) Reasign(w http.ResponseWriter, r *http.Request) error{
    phone := r.URL.Query().Get("phone")
    if err != nil {
        return err
    }

    leads, err := h.Store.GetLeads(phone)
    if len(leads) == 0 || err{
        return fmt.Errorf("No se encontraron leads para el asesor %s. \nerr: %s", phone, err)
    }
    
    asesores, err := h.Store.GetAllExcept(phone)
    if err != nil {
        return err
    }

    each := math.Floor(len(leads) / len(asesores))
    for i, asesor := range asesores{
        query := `
            UPDATE FROM Leads
            SET asesor_phone = ?
            WHERE phone IN (?)
        `

        query, args, err := h.Store.Db.In(query, asesor.Phone, leads[i:i+each])
        query = h.Store.Db.Rebind(query)
        err = h.Store.Db.Query(query, args...)
        if err != nil{
            return fmt.Errorf("No se pudo reasignar los leads err: %s", err)
        }
    }
}
*/

func (h *AsesorHandler) GetAll(w http.ResponseWriter, r *http.Request) error{
    asesores, err := h.Store.GetAll()
    if err != nil {
        return err
    }

    dataResponse(w, r, asesores)
    return nil
}

func (h *AsesorHandler) GetOne(w http.ResponseWriter, r *http.Request) error{
    phone := mux.Vars(r)["phone"]

    asesor, err := h.Store.GetOne(phone)
    if err != nil {
        return err
    }

    dataResponse(w, r, asesor)
    return nil
}

func (h *AsesorHandler) Insert(w http.ResponseWriter, r *http.Request) error{
    var asesor models.Asesor
    reqBody, err := ioutil.ReadAll(r.Body)
    if err != nil {
        return err
    }

    err = json.Unmarshal(reqBody, &asesor)
    if err != nil{
        return err
    }

    validate := validator.New()
    err = validate.Struct(asesor)
    if err != nil{
        return err
    }

    _, err = h.Store.Insert(&asesor)
    if err != nil{
        return err
    }

    successResponse(w, r, "Asesor creado correctamente", asesor)
    return nil
}

func (h *AsesorHandler) Update(w http.ResponseWriter, r *http.Request) error{
    phone := mux.Vars(r)["phone"]

    var updateAsesor models.UpdateAsesor
    reqBody, err := ioutil.ReadAll(r.Body)
    if err != nil {
        return err
    }

    if err = json.Unmarshal(reqBody, &updateAsesor); err != nil{
        return err
    }
    asesor := models.Asesor{
        Phone: phone,
        Name: updateAsesor.Name,
        Active: updateAsesor.Active,
    }

    validate := validator.New()
    if err = validate.Struct(asesor); err != nil{
        return err
    }

    err = h.Store.Update(&asesor, phone)
    if err != nil{
        return err
    }

    successResponse(w, r, "Asesor actualizado correctamente", asesor)
    return nil
}
