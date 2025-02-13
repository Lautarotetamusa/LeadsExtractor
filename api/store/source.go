package store

import (
	"database/sql"
	"fmt"
	"leadsextractor/models"
	"slices"
)

var validSources = []string{"whatsapp", "ivr", "viewphone", "inmuebles24", "lamudi", "casasyterrenos", "propiedades"}

func isValidSource(source string) bool {
	return slices.Contains(validSources, source)
}

func (s *Store) GetSource(c *models.Communication) (*models.Source, error) {
	source := models.Source{}
    if !isValidSource(c.Fuente) {
		return nil, fmt.Errorf("la fuente %s es incorrecta, debe ser (whatsapp, ivr, inmuebles24, lamudi, casasyterrenos, propiedades)", c.Fuente)
    }

	if slices.Contains([]string{"whatsapp", "ivr", "viewphone"}, c.Fuente) {
		err := s.db.Get(&source, "SELECT * FROM Source WHERE type LIKE ?", c.Fuente)
		if err != nil {
			return nil, fmt.Errorf("source: %s no existe", c.Fuente)
		}
		return &source, nil
	}

	property, err := s.insertOrGetProperty(c)
	if err != nil {
		return nil, err
	}

	err = s.db.Get(&source, "SELECT * FROM Source WHERE property_id=?", property.ID.Int32)
	if err != nil {
        return nil, fmt.Errorf("error obteniendo el source con property_id: %d", property.ID.Int32)
	}
	return &source, nil
}

func (s *Store) insertSource(propertyId int) error {
    id := sql.NullInt16{Int16: int16(propertyId), Valid: true}
    if !id.Valid {
        return fmt.Errorf("el property_id %d no es valido", propertyId)
    }
    source := models.Source{
        Tipo:       "property",
        PropertyId: id,
    }
    query := "INSERT INTO Source (type, property_id) VALUES(:type, :property_id)"
    if _, err := s.db.NamedExec(query, source); err != nil {
        return fmt.Errorf("error insertando source: %s", err.Error())
    }
    return nil
}
