package store

import (
	"database/sql"
	"fmt"
	"leadsextractor/models"
	"slices"
)

func (s *Store) GetSource(c *models.Communication) (*models.Source, error) {
	source := models.Source{}
	validSources := []string{"whatsapp", "ivr", "viewphone", "inmuebles24", "lamudi", "casasyterrenos", "propiedades"}
	if !slices.Contains(validSources, c.Fuente) {
		return nil, fmt.Errorf("la fuente %s es incorrecta, debe ser (whatsapp, ivr, inmuebles24, lamudi, casasyterrenos, propiedades)", c.Fuente)
	}

	if slices.Contains([]string{"whatsapp", "ivr", "viewphone"}, c.Fuente) {
		err := s.Db.Get(&source, "SELECT * FROM Source WHERE type LIKE ?", c.Fuente)
		if err != nil {
			return nil, fmt.Errorf("source: %s no existe", c.Fuente)
		}
		return &source, nil
	}

	property, err := s.insertOrGetProperty(c)
	if err != nil {
		return nil, err
	}

	err = s.Db.Get(&source, "SELECT * FROM Source WHERE property_id=?", property.ID.Int32)
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
    if _, err := s.Db.NamedExec(query, source); err != nil {
        return fmt.Errorf("error insertando source: %s", err.Error())
    }
    return nil
}

func (s *Store) insertProperty(c *models.Communication, p *models.Propiedad) error {
    query := "INSERT INTO Property (portal_id, title, url, price, ubication, tipo, portal) VALUES (:portal_id, :title, :url, :price, :ubication, :tipo, :portal)"
    property := c.Propiedad
    property.Portal = c.Fuente

    if _, err := s.Db.NamedExec(query, &property); err != nil {
        return fmt.Errorf("error insertando propiedad: %s", err.Error())
    }

    query = "SELECT * FROM Property WHERE portal_id LIKE ? AND portal = ? LIMIT 1"
    err := s.Db.Get(p, query, c.Propiedad.PortalId, c.Fuente)
    if err != nil {
        return fmt.Errorf("error encontrando la propiedad: %s", err.Error())
    }
    return nil
}

func (s *Store) insertOrGetProperty(c *models.Communication) (*models.Propiedad, error) {
	property := models.Propiedad{}
	query := "SELECT * FROM Property WHERE portal_id LIKE ? AND portal = ? LIMIT 1"
	err := s.Db.Get(&property, query, c.Propiedad.ID.Int32, c.Fuente)

	if err == sql.ErrNoRows {
        err = s.insertProperty(c, &property)
        if err != nil {
            return nil, err
        }
        err = s.insertSource(int(property.ID.Int32))
        if err != nil {
            return nil, err
        }
	}

	return &property, nil
}
