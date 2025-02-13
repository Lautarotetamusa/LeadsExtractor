package store

import (
	"database/sql"
	"fmt"
	"leadsextractor/models"
)

func (s *Store) insertProperty(c *models.Communication, p *models.Propiedad) error {
    query := "INSERT INTO Property (portal_id, title, url, price, ubication, tipo, portal) VALUES (:portal_id, :title, :url, :price, :ubication, :tipo, :portal)"
    property := c.Propiedad
    property.Portal = c.Fuente

    if _, err := s.db.NamedExec(query, &property); err != nil {
        return fmt.Errorf("error insertando propiedad: %s", err.Error())
    }

    query = "SELECT * FROM Property WHERE portal_id LIKE ? AND portal = ? LIMIT 1"
    err := s.db.Get(p, query, c.Propiedad.PortalId, c.Fuente)
    if err != nil {
        return fmt.Errorf("error encontrando la propiedad: %s", err.Error())
    }
    return nil
}

func (s *Store) insertOrGetProperty(c *models.Communication) (*models.Propiedad, error) {
	property := models.Propiedad{}
	query := "SELECT * FROM Property WHERE portal_id LIKE ? AND portal = ? LIMIT 1"
	err := s.db.Get(&property, query, c.Propiedad.ID.Int32, c.Fuente)

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
