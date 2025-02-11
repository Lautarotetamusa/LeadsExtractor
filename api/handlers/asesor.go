package handlers

import (
	"encoding/json"
	"fmt"

	"leadsextractor/models"
	"leadsextractor/pkg/roundrobin"
	"leadsextractor/store"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type AsesorHandler struct {
    service *AsesorService
}

type AsesorService struct {
    roundRobin  *roundrobin.RoundRobin[models.Asesor]

    asesor      store.AsesorStorer
    lead        store.LeadStorer
}

func NewAsesorHandler(s *AsesorService) *AsesorHandler {
    return &AsesorHandler{
        service: s,
    }
}

func NewAsesorService(asesor store.AsesorStorer, lead store.LeadStorer, rr *roundrobin.RoundRobin[models.Asesor]) *AsesorService {
    return &AsesorService{
        asesor: asesor,
        lead: lead,
        roundRobin: rr,
    }
}

func (h *AsesorHandler) RegisterRoutes(router *mux.Router) {
    r := router.PathPrefix("/asesor").Subrouter()

	r.HandleFunc("", HandleErrors(h.GetAll)).Methods(http.MethodGet)
	r.HandleFunc("/{phone}", HandleErrors(h.GetOne)).Methods(http.MethodGet)
	r.HandleFunc("", HandleErrors(h.Insert)).Methods(http.MethodPost)
	r.HandleFunc("/{phone}", HandleErrors(h.Update)).Methods(http.MethodPut)
	r.HandleFunc("/{phone}", HandleErrors(h.Delete)).Methods(http.MethodDelete)
	r.HandleFunc("/{phone}", HandleErrors(h.Delete)).Methods(http.MethodDelete)
    r.HandleFunc("/{phone}/reasign", HandleErrors(h.Reasign)).Methods(http.MethodPut)
    r.HandleFunc("/{phone}/leads", HandleErrors(h.GetLeads)).Methods(http.MethodGet)
}

// Reasigns all the leads of an asesor to all the others active asesores with the round-robin method.
func (s *AsesorService) ReasignLeads(a *models.Asesor) (int, error) {
    a.Active = false
    if err := s.asesor.Update(a); err != nil {
        return 0, fmt.Errorf("update asesor failed")
    }

    asesores, err := s.asesor.GetAllActive()
    if err != nil {
        return 0, fmt.Errorf("impossible to get asesores list")
    }
    if len(asesores) == 0 {
        return 0, ErrBadRequest("all the asesores are inactive")
    }
    s.roundRobin.Reasign(asesores)

    leads, err := s.asesor.GetLeads(a.Phone.String())
    if err != nil {
        return 0, fmt.Errorf("no fue posible obtener los leads del asesor")
    }

    // Update all the leads of that asesor
    // TODO: Goroutines
    for _, lead := range leads {
        nextAsesor := s.roundRobin.Next()
        if err = s.lead.UpdateAsesor(lead.Phone, nextAsesor); err != nil {
            return 0, fmt.Errorf("no fue posible reasignar a %s", lead.Phone)
        }
    }
    return len(leads), nil
}

func (h *AsesorHandler) GetAll(w http.ResponseWriter, r *http.Request) error {
    asesores, err := h.service.asesor.GetAll()
	if err != nil {
		return err
	}

	dataResponse(w, asesores)
	return nil
}

func (h *AsesorHandler) GetOne(w http.ResponseWriter, r *http.Request) error {
	phone := mux.Vars(r)["phone"]

	asesor, err := h.service.asesor.GetOne(phone)
	if err != nil {
        return err
	}

	dataResponse(w, asesor)
	return nil
}

func (h *AsesorHandler) GetLeads(w http.ResponseWriter, r *http.Request) error {
	phone := mux.Vars(r)["phone"]

	leads, err := h.service.asesor.GetLeads(phone)
	if err != nil {
        return err
	}

	dataResponse(w, leads)
	return nil
}

func (h *AsesorHandler) Delete(w http.ResponseWriter, r *http.Request) error {
    phone := mux.Vars(r)["phone"]
	asesor, err := h.service.asesor.GetOne(phone)
	if err != nil {
        return err
	}

    if err := h.service.asesor.Delete(asesor); err != nil {
		return err
	}

    messageResponse(w, "Asesor eliminado con exito")
	return nil
}

func (h *AsesorHandler) Insert(w http.ResponseWriter, r *http.Request) error {
	var asesor models.Asesor
    defer r.Body.Close()
    if err := json.NewDecoder(r.Body).Decode(&asesor); err != nil {
        return jsonErr(err)
    }

	validate := validator.New()
    if err := validate.Struct(asesor); err != nil {
		return ErrBadRequest(err.Error())
	}

    if err := h.service.asesor.Insert(&asesor); err != nil {
		return err
	}

    if asesor.Active {
        h.service.roundRobin.Add(&asesor)
    }

	createdResponse(w, "Asesor creado correctamente", asesor)
	return nil
}

func (h *AsesorHandler) Reasign(w http.ResponseWriter, r *http.Request) error {
    phone := mux.Vars(r)["phone"]

	asesor, err := h.service.asesor.GetOne(phone)
	if err != nil {
        return err
	}

    n, err := h.service.ReasignLeads(asesor)
    if err != nil {
        return err
    }

    messageResponse(w, fmt.Sprintf("se reasignaron un total de %d leads", n))
    return nil
}

func (h *AsesorHandler) Update(w http.ResponseWriter, r *http.Request) error {
	phone := mux.Vars(r)["phone"]

	var updateAsesor models.UpdateAsesor
    defer r.Body.Close()
    if err := json.NewDecoder(r.Body).Decode(&updateAsesor); err != nil {
		return jsonErr(err)
    }

    asesor, err := h.service.asesor.GetOne(phone)
    if err != nil {
        return err
    }

    updateFields(asesor, updateAsesor)
    if err := h.service.asesor.Update(asesor); err != nil {
		return err
	}

    // if active field was updated, then assign all the active asesores to the round-robin
    if updateAsesor.Active != nil {
        asesores, err := h.service.asesor.GetAllActive()
        if err != nil {
            return fmt.Errorf("error getting the list of asesores")
        }
        h.service.roundRobin.Reasign(asesores)
    }

	createdResponse(w, "Asesor actualizado correctamente", asesor)
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
