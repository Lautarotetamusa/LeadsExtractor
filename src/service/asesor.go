package service

import (
	"fmt"

	"leadsextractor/models"
	"leadsextractor/pkg/roundrobin"
	"leadsextractor/store"

	"github.com/go-playground/validator/v10"
)

type AsesorService struct {
	validate   *validator.Validate
	roundRobin *roundrobin.RoundRobin[models.Asesor]

	asesor store.AsesorStorer
	lead   store.LeadStorer
}

func NewAsesorService(asesor store.AsesorStorer, lead store.LeadStorer, rr *roundrobin.RoundRobin[models.Asesor], validate *validator.Validate) *AsesorService {
	return &AsesorService{
		asesor:     asesor,
		lead:       lead,
		roundRobin: rr,
		validate:   validate,
	}
}

func (s *AsesorService) GetAll() ([]*models.Asesor, error) {
	return s.asesor.GetAll()
}

func (s *AsesorService) GetOne(phone string) (*models.Asesor, error) {
	return s.asesor.GetOne(phone)
}

func (s *AsesorService) GetLeads(phone string) ([]*models.Lead, error) {
	return s.asesor.GetLeads(phone)
}

func (s *AsesorService) Insert(asesor *models.Asesor) error {
	if err := s.validate.Struct(asesor); err != nil {
		return NewValidationError(err.Error())
	}

	if err := s.asesor.Insert(asesor); err != nil {
		return err
	}

	if asesor.Active {
		s.roundRobin.Add(asesor)
	}
	return nil
}

func (s *AsesorService) Update(phone string, update models.UpdateAsesor) (*models.Asesor, error) {
	asesor, err := s.asesor.GetOne(phone)
	if err != nil {
		return nil, err
	}

	updateAsesorFields(asesor, update)
	if err := s.asesor.Update(asesor); err != nil {
		return nil, err
	}

	// if active field was updated, then assign all the active asesores to the round-robin
	if update.Active != nil {
		asesores, err := s.asesor.GetAllActive()
		if err != nil {
			return nil, fmt.Errorf("error getting the list of asesores")
		}
		s.roundRobin.Reasign(asesores)
	}

	return asesor, nil
}

func (s *AsesorService) Delete(phone string) error {
	asesor, err := s.asesor.GetOne(phone)
	if err != nil {
		return err
	}
	return s.asesor.Delete(asesor)
}

// ReasignLeads desactiva al asesor y redistribuye todos sus leads entre los
// demás asesores activos con el round-robin.
func (s *AsesorService) ReasignLeads(phone string) (int, error) {
	asesor, err := s.asesor.GetOne(phone)
	if err != nil {
		return 0, err
	}

	asesor.Active = false
	if err := s.asesor.Update(asesor); err != nil {
		return 0, fmt.Errorf("update asesor failed")
	}

	asesores, err := s.asesor.GetAllActive()
	if err != nil {
		return 0, fmt.Errorf("impossible to get asesores list")
	}
	if len(asesores) == 0 {
		return 0, NewValidationError("all the asesores are inactive")
	}
	s.roundRobin.Reasign(asesores)

	leads, err := s.asesor.GetLeads(asesor.Phone.String())
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

func updateAsesorFields(asesor *models.Asesor, update models.UpdateAsesor) {
	if update.Name != nil {
		asesor.Name = *update.Name
	}
	if update.Phone != nil {
		asesor.Phone = *update.Phone
	}
	if update.Email != nil {
		asesor.Email = *update.Email
	}
	if update.Active != nil {
		asesor.Active = *update.Active
	}
}
