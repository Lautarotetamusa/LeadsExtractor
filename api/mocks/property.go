package mocks

import (
	"fmt"
	"leadsextractor/models"
	"leadsextractor/store"
	"time"
)

type MockPropertyStorer struct {
	props []*store.PortalProp
    images []store.PropertyImage
}

var lastImgId = 0

func (s *MockPropertyStorer) Mock() {
    mockData := []*store.PortalProp{
        {
            ID:            1,
            Title:         "Beautiful Apartment in Downtown",
            Price:         "500000",
            Currency:      "USD",
            Description:   "A spacious 3-bedroom apartment with a stunning view.",
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
            Description:   "A charming 2-bedroom house with a large garden.",
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
            Description:   "A modern loft with high ceilings and large windows.",
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


func (s *MockPropertyStorer) InsertImages(propId int64, images []store.PropertyImage) error {
    for _, image := range images {
        image.PropertyID = propId
        image.ID = int64(lastImgId + 1) 
        lastImgId++
        s.images = append(s.images, image)
    }

    _, err := s.GetOne(propId)
    return err
}

func (s *MockPropertyStorer) GetImages(p *store.PortalProp) error {
    p.Images = make([]store.PropertyImage, 0)
    for _, image := range s.images {
        if image.PropertyID == p.ID {
            p.Images = append(p.Images, image)
        }
    }

    fmt.Printf("PROP IMAGES: %#v\n", p.Images)
    _, err := s.GetOne(p.ID)
    return err
}

func (s *MockPropertyStorer) DeleteImage(propId, imageId int64) error {
    exists := false
    images := make([]store.PropertyImage, 0)
    for _, image := range s.images {
        if image.ID == imageId {
            exists = true
            continue
        }
        images = append(images, image)
    }
    if !exists {
        return store.NewErr("image does not exists", store.StoreNotFoundErr)
    }

    s.images = images

    _, err := s.GetOne(propId)
    return err
}

func (s *MockPropertyStorer) Insert(prop *store.PortalProp) error {
    prop.ID = int64(len(s.props) + 1)
	s.props = append(s.props, prop)
    s.InsertImages(prop.ID, prop.Images)
    s.GetImages(prop)

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
