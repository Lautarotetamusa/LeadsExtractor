package pkg

import (
	"encoding/json"

	"leadsextractor/models"
	"leadsextractor/store"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type AsesorHandler struct {
	Store *store.Store
}

func (s *Server) UpdateStatuses(w http.ResponseWriter, r *http.Request) error {
	var asesores []models.Asesor
	defer r.Body.Close()
    if err := json.NewDecoder(r.Body).Decode(&asesores); err != nil {
		return err
    }

	for i := range asesores {
        if err := s.Store.UpdateAsesor(&asesores[i], asesores[i].Phone); err != nil {
			return err
		}
	}

    if err := s.Store.GetAllActiveAsesores(&asesores); err != nil{
        return err
    }

	s.roundRobin.SetAsesores(asesores)

	successResponse(w, "Asesores actualizados correctamente", nil)
	return nil
}

func (s *Server) AssignAsesor(w http.ResponseWriter, r *http.Request) error {
	c := &models.Communication{} //NO poner como var porque el Decode deja de funcionar
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(c); err != nil {
		return err
	}

	lead, err := s.Store.InsertOrGetLead(s.roundRobin, c)
	if err != nil {
		return err
	}
	c.Asesor = lead.Asesor

	w.Header().Set("Content-Type", "application/json")
	res := struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
		IsNew   bool        `json:"is_new"`
	}{true, c, c.IsNew}

	json.NewEncoder(w).Encode(res)
	return nil
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

func (s *Server) UpdateAsesor(w http.ResponseWriter, r *http.Request) error {
	phone := mux.Vars(r)["phone"]

	var updateAsesor models.UpdateAsesor
    defer r.Body.Close()
    if err := json.NewDecoder(r.Body).Decode(&updateAsesor); err != nil {
		return err
    }

	asesor := models.Asesor{
		Phone:  phone,
		Name:   updateAsesor.Name,
		Active: updateAsesor.Active,
	}

	validate := validator.New()
    if err := validate.Struct(asesor); err != nil {
		return err
	}

    if err := s.Store.UpdateAsesor(&asesor, phone); err != nil {
		return err
	}

	successResponse(w, "Asesor actualizado correctamente", asesor)
	return nil
}
