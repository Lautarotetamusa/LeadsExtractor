package pkg

import (
	"encoding/json"
	"io"

	"leadsextractor/models"
	"leadsextractor/store"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type ReasignData struct {
	Phone string `json:"phone"`
}

type AsesorHandler struct {
	Store *store.Store
}

func (s *Server) UpdateStatuses(w http.ResponseWriter, r *http.Request) error {
	var asesores []models.Asesor

	defer r.Body.Close()
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(reqBody, &asesores)
	if err != nil {
		return err
	}

	for i := range asesores {
		err = s.Store.UpdateAsesor(&asesores[i], asesores[i].Phone)
		if err != nil {
			return err
		}
	}

    err = s.Store.GetAllActiveAsesores(&asesores)

    if err != nil{
        return err
    }

	s.roundRobin.SetAsesores(asesores)

	successResponse(w, r, "Asesores actualizados correctamente", nil)
	return nil
}

func (s *Server) AssignAsesor(w http.ResponseWriter, r *http.Request) error {
	c := &models.Communication{}
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(c)
	if err != nil {
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

func (s *Server) GetAllAsesores(w http.ResponseWriter, r *http.Request) error {
	asesores := []models.Asesor{}
	err := s.Store.GetAllAsesores(&asesores)
	if err != nil {
		return err
	}

	dataResponse(w, r, asesores)
	return nil
}

func (s *Server) GetOneAsesor(w http.ResponseWriter, r *http.Request) error {
	phone := mux.Vars(r)["phone"]

	asesor, err := s.Store.GetOneAsesor(phone)
	if err != nil {
		return err
	}

	dataResponse(w, r, asesor)
	return nil
}

func (s *Server) InsertAsesor(w http.ResponseWriter, r *http.Request) error {
	var asesor models.Asesor
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(reqBody, &asesor)
	if err != nil {
		return err
	}

	validate := validator.New()
	err = validate.Struct(asesor)
	if err != nil {
		return err
	}

	_, err = s.Store.InsertAsesor(&asesor)
	if err != nil {
		return err
	}

	successResponse(w, r, "Asesor creado correctamente", asesor)
	return nil
}

func (s *Server) UpdateAsesor(w http.ResponseWriter, r *http.Request) error {
	phone := mux.Vars(r)["phone"]

	var updateAsesor models.UpdateAsesor
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(reqBody, &updateAsesor); err != nil {
		return err
	}
	asesor := models.Asesor{
		Phone:  phone,
		Name:   updateAsesor.Name,
		Active: updateAsesor.Active,
	}

	validate := validator.New()
	if err = validate.Struct(asesor); err != nil {
		return err
	}

	err = s.Store.UpdateAsesor(&asesor, phone)
	if err != nil {
		return err
	}

	successResponse(w, r, "Asesor actualizado correctamente", asesor)
	return nil
}
