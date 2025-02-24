package mocks

import (
	"leadsextractor/models"
	"leadsextractor/store"
	"time"
)

type MockPropertyStorer struct {
	props []*store.PortalProp
}

func (s *MockPropertyStorer) Mock() {
    println("MOCKING PROPS..")
    mockData := []*store.PortalProp{
        {
            ID:            1,
            Title:         "Beautiful Apartment in Downtown",
            Price:         "500000",
            Currency:      "USD",
            Description:   models.NullString{String: "A spacious 3-bedroom apartment with a stunning view.", Valid: true},
            Type:          models.NullString{String: "Apartment", Valid: true},
            Antiquity:     5,
            ParkingLots:   models.NullInt16{Int16: 2, Valid: true},
            Bathrooms:     models.NullInt16{Int16: 2, Valid: true},
            HalfBathrooms: models.NullInt16{Int16: 1, Valid: true},
            Rooms:         models.NullInt16{Int16: 3, Valid: true},
            OperationType: "sale",
            M2Total:       models.NullInt16{Int16: 150, Valid: true},
            M2Covered:     models.NullInt16{Int16: 120, Valid: true},
            VideoURL:      models.NullString{String: "https://example.com/video1", Valid: true},
            VirtualRoute:  models.NullString{String: "https://example.com/virtual-tour1", Valid: true},
            UpdatedAt:     time.Now(),
            CreatedAt:     time.Now(),
        },
        {
            ID:            2,
            Title:         "Cozy House in the Suburbs",
            Price:         "300000",
            Currency:      "EUR",
            Description:   models.NullString{String: "A charming 2-bedroom house with a large garden.", Valid: true},
            Type:          models.NullString{String: "House", Valid: true},
            Antiquity:     10,
            ParkingLots:   models.NullInt16{Int16: 1, Valid: true},
            Bathrooms:     models.NullInt16{Int16: 1, Valid: true},
            HalfBathrooms: models.NullInt16{Int16: 0, Valid: true},
            Rooms:         models.NullInt16{Int16: 2, Valid: true},
            OperationType: "rent",
            M2Total:       models.NullInt16{Int16: 200, Valid: true},
            M2Covered:     models.NullInt16{Int16: 150, Valid: true},
            VideoURL:      models.NullString{String: "", Valid: false},
            VirtualRoute:  models.NullString{String: "", Valid: false},
            UpdatedAt:     time.Now(),
            CreatedAt:     time.Now(),
        },
        {
            ID:            3,
            Title:         "Modern Loft in the City Center",
            Price:         "750000",
            Currency:      "USD",
            Description:   models.NullString{String: "A modern loft with high ceilings and large windows.", Valid: true},
            Type:          models.NullString{String: "Loft", Valid: true},
            Antiquity:     2,
            ParkingLots:   models.NullInt16{Int16: 1, Valid: true},
            Bathrooms:     models.NullInt16{Int16: 1, Valid: true},
            // HalfBathrooms: models.NullInt16{Int16: 1, Valid: true},
            // Rooms:         models.NullInt16{Int16: 1, Valid: true},
            OperationType: "sale",
            M2Total:       models.NullInt16{Int16: 100, Valid: true},
            M2Covered:     models.NullInt16{Int16: 80, Valid: true},
            VideoURL:      models.NullString{String: "https://example.com/video2", Valid: true},
            VirtualRoute:  models.NullString{String: "https://example.com/virtual-tour2", Valid: true},
            UpdatedAt:     time.Now(),
            CreatedAt:     time.Now(),
        },
    }

    for _, prop := range mockData {
        s.Insert(prop)
    }
}

func (s *MockPropertyStorer) Props() []*store.PortalProp {
	return s.props
}

func (s *MockPropertyStorer) GetAll() ([]*store.PortalProp, error) {
	return s.props, nil
}

func (s *MockPropertyStorer) GetOne(id int64) (*store.PortalProp, error) {
	for _, prop := range s.props {
		if prop.ID == id {
			return prop, nil
		}
	}
	return nil, store.NewErr("does not exists", store.StoreNotFoundErr)
}

func (s *MockPropertyStorer) Insert(prop *store.PortalProp) error {
	s.props = append(s.props, prop)
    prop.ID = int64(len(s.props))

    prop.UpdatedAt = time.Time{}
    prop.UpdatedAt = time.Time{}

	return nil
}

func (s *MockPropertyStorer) Update(uProp *store.PortalProp) error {
	for i, prop := range s.props {
		if prop.ID == uProp.ID {
			s.props[i] = uProp
			return nil
		}
	}
	return store.NewErr("not found", store.StoreNotFoundErr)
}
