package mocks

import (
	"leadsextractor/models"
	"leadsextractor/pkg/numbers"
	"leadsextractor/store"
)

type MockAsesorStorer struct {
	asesores []*models.Asesor
	leads    map[string][]*models.Lead
}

func NewMockAsesorStore() *MockAsesorStorer {
	return &MockAsesorStorer{
		asesores: make([]*models.Asesor, 0),
		leads:    make(map[string][]*models.Lead),
	}
}

func (s *MockAsesorStorer) Asesores() []*models.Asesor {
	return s.asesores
}

func (s *MockAsesorStorer) Leads() map[string][]*models.Lead {
	return s.leads
}

func (s *MockAsesorStorer) Mock() {
	// Name   string `db:"name"   json:"name" validate:"required"`
	// Phone  numbers.PhoneNumber `db:"phone"  json:"phone" validate:"required"`
	// Email  string `db:"email"  json:"email" validate:"required"`
	// Active bool   `db:"active" json:"active" validate:"required"`
	s.asesores = []*models.Asesor{
		{"juan", numbers.PhoneNumber("5493415554444"), "juan@gmail.com", true},
		{"carlos", numbers.PhoneNumber("5493415556666"), "carlos@gmail.com", true},
		{"jose", numbers.PhoneNumber("5493415557777"), "jose@gmail.com", true},
		{"maria", numbers.PhoneNumber("5493415558888"), "maria@gmail.com", false},
		{"juana", numbers.PhoneNumber("5493415559999"), "juana@gmail.com", true},
	}

	email := models.NullString{String: "lead@gmail.com", Valid: true}
	s.leads["5493415554444"] = []*models.Lead{
		{"lead1", numbers.PhoneNumber("5493414449999"), email, *s.asesores[0], ""},
		{"lead2", numbers.PhoneNumber("5493414448888"), email, *s.asesores[0], ""},
		{"lead3", numbers.PhoneNumber("5493414447777"), email, *s.asesores[0], ""},
	}
	s.leads["5493415559999"] = []*models.Lead{
		{"lead1", numbers.PhoneNumber("5493414449999"), email, *s.asesores[4], ""},
		{"lead2", numbers.PhoneNumber("5493414448888"), email, *s.asesores[4], ""},
		{"lead3", numbers.PhoneNumber("5493414447777"), email, *s.asesores[4], ""},
	}
}

func (s *MockAsesorStorer) GetAll() ([]*models.Asesor, error) {
	return s.asesores, nil
}

func (s *MockAsesorStorer) GetAllActive() ([]*models.Asesor, error) {
	return s.asesores, nil
}

func (s *MockAsesorStorer) GetLeads(phone string) ([]*models.Lead, error) {
	a, _ := s.GetOne(phone)
	leads := s.leads[a.Phone.String()]
	return leads, nil
}

func (s *MockAsesorStorer) GetOne(phone string) (*models.Asesor, error) {
	for _, a := range s.asesores {
		if a.Phone.String() == phone {
			return a, nil
		}
	}
	return nil, store.NewErr("does not exists", store.StoreNotFoundErr)
}

func (s *MockAsesorStorer) GetFromEmail(email string) (*models.Asesor, error) {
	for _, a := range s.asesores {
		if a.Email == email {
			return a, nil
		}
	}
	return nil, store.NewErr("not found", store.StoreNotFoundErr)
}

func (s *MockAsesorStorer) Insert(asesor *models.Asesor) error {
	if l, _ := s.GetOne(asesor.Phone.String()); l != nil {
		return store.NewErr("already exists", store.StoreDuplicatedErr)
	}

	s.asesores = append(s.asesores, asesor)
	return nil
}

func (s *MockAsesorStorer) Update(asesor *models.Asesor) error {
	for i, a := range s.asesores {
		if a.Phone.String() == asesor.Phone.String() {
			s.asesores[i] = asesor
			return nil
		}
	}
	return store.NewErr("not found", store.StoreNotFoundErr)
}

func (s *MockAsesorStorer) Delete(asesor *models.Asesor) error {
	return nil
}
