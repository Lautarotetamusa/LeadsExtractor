package mocks

import (
	"errors"
	"leadsextractor/models"
)

type MockPropertyStorer struct {
	properties []*models.Propiedad
}

func (s *MockPropertyStorer) InsertProperty(p *models.Propiedad) error {
	p.ID = models.NullInt32{Int32: int32(len(s.properties) + 1), Valid: true}
	s.properties = append(s.properties, p)
	return nil
}

func (s *MockPropertyStorer) GetProperty(portalId string, portal string) (*models.Propiedad, error) {
	for _, prop := range s.properties {
		if prop.PortalId.String == portalId && prop.Portal == portal {
			return prop, nil
		}
	}
	return nil, errors.New("property not found")
}
