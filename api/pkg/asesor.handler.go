package pkg

import (
	"encoding/json"
	"fmt"

	"leadsextractor/models"
	"leadsextractor/store"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type AsesorHandler struct {
	Store *store.Store
}

func (s *Server) GetAllAsesores(w http.ResponseWriter, r *http.Request) error {
    var asesores []models.Asesor
	if err := s.Store.GetAllAsesores(&asesores); err != nil {
		return err
	}

	dataResponse(w, asesores)
	return nil
}

func (s *Server) GetOneAsesor(w http.ResponseWriter, r *http.Request) error {
	phone := mux.Vars(r)["phone"]

	asesor, err := s.Store.GetOneAsesor(phone)
	if err != nil {
		return err
	}

	dataResponse(w, asesor)
	return nil
}

func (s *Server) InsertAsesor(w http.ResponseWriter, r *http.Request) error {
	var asesor models.Asesor
    defer r.Body.Close()
    if err := json.NewDecoder(r.Body).Decode(&asesor); err != nil {
        return err
    }

	validate := validator.New()
    if err := validate.Struct(asesor); err != nil {
		return err
	}

    if err := s.Store.InsertAsesor(&asesor); err != nil {
		return err
	}

	successResponse(w, "Asesor creado correctamente", asesor)
	return nil
}

func (s *Server) Reasign(w http.ResponseWriter, r *http.Request) error {
    phone := mux.Vars(r)["phone"]

    asesorReasignado, err := s.Store.GetOneAsesor(phone)
    if err != nil {
        return fmt.Errorf("no se encontro el asesor con telefono %s", phone)
    }
    asesorReasignado.Active = false

    if err = s.Store.UpdateAsesor(asesorReasignado.Phone, asesorReasignado); err != nil {
        return fmt.Errorf("no fue posible actualizar al asesor")
    }

    var asesores []models.Asesor
    if err = s.Store.GetAllActiveAsesores(&asesores); err != nil {
        return fmt.Errorf("no fue posible obtener la lista de asesores")
    }
    s.roundRobin.SetAsesores(asesores)

    leads, err := s.Store.GetLeadsFromAsesor(phone)
    if err != nil {
        return fmt.Errorf("no fue posible obtener los leads del asesor")
    }
    for _, lead := range *leads{
        if err = s.Store.UpdateLeadAsesor(lead.Phone, asesorReasignado); err != nil {
            return fmt.Errorf("no fue posible reasignar a %s", lead.Phone)
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{
        "success": true,
        "message": fmt.Sprintf("se reasignaron un total de %d leads", len(*leads)),
    })
    return nil
}

func (s *Server) UpdateAsesor(w http.ResponseWriter, r *http.Request) error {
	phone := mux.Vars(r)["phone"]

	var updateAsesor models.UpdateAsesor
    defer r.Body.Close()
    if err := json.NewDecoder(r.Body).Decode(&updateAsesor); err != nil {
		return err
    }

    asesor, err := s.Store.GetOneAsesor(phone)
    if err != nil {
        return fmt.Errorf("no existe asesor con telefono %s", phone)
    }

    updateFields(asesor, updateAsesor)

    if err := s.Store.UpdateAsesor(phone, asesor); err != nil {
		return err
	}

    if updateAsesor.Active != nil {
        s.logger.Debug("Actualizando round robin")
        var asesores []models.Asesor
        err := s.Store.GetAllActiveAsesores(&asesores)
        if err != nil {
            return fmt.Errorf("no fue posible obtener la lista de asesores")
        }
        s.roundRobin.SetAsesores(asesores)
    }

	successResponse(w, "Asesor actualizado correctamente", asesor)
	return nil
}

func updateFields(asesor *models.Asesor, updateAsesor models.UpdateAsesor) {
	if updateAsesor.Name != nil {
		asesor.Name = *updateAsesor.Name
	}
	if updateAsesor.Phone != nil {
		asesor.Phone = *updateAsesor.Phone
	}
	if updateAsesor.Email != nil {
		asesor.Email = *updateAsesor.Email
	}
	if updateAsesor.Active != nil {
		asesor.Active = *updateAsesor.Active
	}
}
