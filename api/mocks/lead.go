package mocks

import (
	"leadsextractor/models"
	"leadsextractor/pkg/numbers"
	"leadsextractor/store"
)

type MockLeadStorer struct {
    leads []models.Lead
}

func (s *MockLeadStorer) Mock() {
    s.Insert(&models.CreateLead{
        Name: "test",
        Phone: numbers.PhoneNumber("5493415854220"),
        Email: models.NullString{String: "test@gmail.com", Valid: true},
        AsesorPhone: numbers.PhoneNumber("5493415854212"),
        Cotizacion: "",
    })
}

func (s *MockLeadStorer) Leads() *[]models.Lead {
    return &s.leads
}

func (s *MockLeadStorer) GetAll() (*[]models.Lead, error) {
    return &s.leads, nil
}

func (s *MockLeadStorer) GetOne(phone numbers.PhoneNumber) (*models.Lead, error) {
    for _, lead := range s.leads {
        if lead.Phone == phone {
            return &lead, nil
        }
    }
    return nil, store.NewErr("does not exists", store.StoreNotFoundErr)
}

func (s *MockLeadStorer) Insert(createLead *models.CreateLead) (*models.Lead, error) {
    if l, _ := s.GetOne(createLead.Phone); l != nil {
        return nil, store.NewErr("already exists", store.StoreDuplicatedErr)
    }

    lead := models.Lead{
        Name: createLead.Name,
        Phone: createLead.Phone,
        Email: createLead.Email,
        Cotizacion: createLead.Cotizacion,
        Asesor: models.Asesor{
            Name: "test",
            Phone: createLead.AsesorPhone,
            Email: "test@gmail.com", // Required field
        },
    }

    s.leads = append(s.leads, lead) 
    return &lead, nil
}

func (s *MockLeadStorer) Update(uLead *models.Lead) error {
    for i, lead := range s.leads {
        if lead.Phone.String() == uLead.Phone.String() {
            s.leads[i] = *uLead
            return nil
        }
    }
    return store.NewErr("not found", store.StoreNotFoundErr)
}

func (s *MockLeadStorer) UpdateAsesor(phone numbers.PhoneNumber, a *models.Asesor) error {
    for i, lead := range s.leads {
        if lead.Phone == phone {
            s.leads[i].Asesor = *a
        }
    }
    return store.NewErr("not found", store.StoreNotFoundErr)
}
