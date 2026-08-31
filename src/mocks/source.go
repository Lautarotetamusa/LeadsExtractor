package mocks

import (
	"database/sql"
	"errors"
	"leadsextractor/models"
)

// type SourceStorer interface {
//     InsertSource(propertyId int) error
//     InsertProperty(*models.Propiedad) error
//     GetSource(sourceType string) (*models.Source, error)
//     GetPropertySource(propertyId int32) (*models.Source, error)
//     GetProperty(portalId string, portal string) (*models.Propiedad, error)
// }

type MockSourceStorer struct {
	sources    []*models.Source
	properties []*models.Propiedad
}

func (s *MockSourceStorer) InsertSource(propertyId int) error {
	s.sources = append(s.sources, &models.Source{
		PropertyId: sql.NullInt16{Int16: int16(propertyId), Valid: true},
		Tipo:       "property",
		Id:         len(s.sources),
	})
	return nil
}

func (s *MockSourceStorer) InsertProperty(p *models.Propiedad) error {
	p.ID = models.NullInt32{Int32: int32(len(s.properties) + 1), Valid: true}
	s.properties = append(s.properties, p)
	return nil
}

func (s *MockSourceStorer) GetSource(sourceType string) (*models.Source, error) {
	for _, source := range s.sources {
		if source.Tipo == sourceType {
			return source, nil
		}
	}
	return nil, errors.New("source type not found")
}

func (s *MockSourceStorer) GetPropertySource(propertyId int32) (*models.Source, error) {
	for _, source := range s.sources {
		if int32(source.PropertyId.Int16) == propertyId {
			return source, nil
		}
	}
	return nil, errors.New("source with prop id not found")
}

func (s *MockSourceStorer) GetProperty(portalId string, portal string) (*models.Propiedad, error) {
	for _, prop := range s.properties {
		if prop.PortalId.String == portalId && prop.Portal == portal {
			return prop, nil
		}
	}
	return nil, errors.New("property not found")
}
