package store

import (
	"database/sql"
	"fmt"
	"leadsextractor/models"
	"log/slog"
	"slices"

	"github.com/jmoiron/sqlx"
)

type Store struct {
	Db *sqlx.DB
    logger *slog.Logger
}

func NewStore(db *sqlx.DB, logger *slog.Logger) *Store {
	return &Store{
		Db: db,
        logger: logger,
	}
}

func (s *Store) GetSource(c *models.Communication) (*models.Source, error) {
	source := models.Source{}
	validSources := []string{"whatsapp", "ivr", "viewphone", "inmuebles24", "lamudi", "casasyterrenos", "propiedades"}
	if !slices.Contains(validSources, c.Fuente) {
		return nil, fmt.Errorf("la fuente %s es incorrecta, debe ser (whatsapp, ivr, inmuebles24, lamudi, casasyterrenos, propiedades)", c.Fuente)
	}

	if c.Fuente == "whatsapp" || c.Fuente == "ivr" || c.Fuente == "viewphone" {
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

	err = s.Db.Get(&source, "SELECT * FROM Source WHERE property_id=?", property.Id)
	if err != nil {
		return nil, err
	}
	return &source, nil
}

func (s *Store) insertProperty(c *models.Communication, p *models.Property) error {
    query := "INSERT INTO Property (portal_id, title, url, price, ubication, tipo, portal) VALUES (:portal_id, :title, :url, :price, :ubication, :tipo, :portal)"
    property := models.Property{
        PortalId:  c.Propiedad.ID,
        Title:     c.Propiedad.Titulo,
        Url:       c.Propiedad.Link,
        Price:     c.Propiedad.Precio,
        Ubication: c.Propiedad.Ubicacion,
        Tipo:      c.Propiedad.Tipo,
        Portal:    c.Fuente,
    }
    if _, err := s.Db.NamedExec(query, &property); err != nil {
        return err
    }

    query = "SELECT * FROM Property WHERE portal_id LIKE ? AND portal = ? LIMIT 1"
    err := s.Db.Get(&p, query, c.Propiedad.ID, c.Fuente)
    if err != nil {
        return err
    }
    return nil
}

func (s *Store) insertSource(propertyId int) error {
    source := models.Source{
        Tipo:       "property",
        PropertyId: sql.NullInt16{Int16: int16(propertyId), Valid: true},
    }
    query := "INSERT INTO Source (type, property_id) VALUES(:type, :property_id)"
    if _, err := s.Db.NamedExec(query, source); err != nil {
        return err
    }
    return nil
}

func (s *Store) insertOrGetProperty(c *models.Communication) (*models.Property, error) {
	property := models.Property{}
	query := "SELECT * FROM Property WHERE portal_id LIKE ? AND portal = ? LIMIT 1"
	err := s.Db.Get(&property, query, c.Propiedad.ID, c.Fuente)

	if err == sql.ErrNoRows {
        slog.Debug("No se encontro property")
    
        s.insertProperty(c, &property)
        s.insertSource(property.Id)
	}

	return &property, nil
}

func (s *Store) InsertCommunication(c *models.Communication, lead *models.Lead, source *models.Source, isNewLead  bool) error {
	query := `INSERT INTO Communication(lead_phone, source_id, new_lead, lead_date, url, zones, mt2_terrain, mt2_builded, baths, rooms) 
    VALUES (:lead_phone, :source_id, :new_lead, :lead_date, :url, :zones, :mt2_terrain, :mt2_builded, :baths, :rooms)`
    _, err := s.Db.NamedExec(query, map[string]interface{}{
		"lead_phone":  lead.Phone,
		"source_id":   source.Id,
		"new_lead":    isNewLead,
		"lead_date":   c.FechaLead,
		"url":         c.Link,
		"zones":       c.Busquedas.Zonas,
		"mt2_terrain": c.Busquedas.TotalArea,
		"mt2_builded": c.Busquedas.CoveredArea,
		"baths":       c.Busquedas.Banios,
		"rooms":       c.Busquedas.Recamaras,
	})
	if err != nil {
		return err
	}
    return nil
}

func (s *Store) InsertOrGetLead(rr *RoundRobin, c *models.Communication) (*models.Lead, bool, error) {
	var isNewLead = false
	var lead *models.Lead

	lead, err := s.GetOne(c.Telefono)

	if err == sql.ErrNoRows {
		isNewLead = true
		c.Asesor = rr.Next()

		lead, err = s.Insert(&models.CreateLead{
			Name:        c.Nombre,
			Phone:       c.Telefono,
			Email:       sql.NullString{String: c.Email},
			AsesorPhone: c.Asesor.Phone,
		})

		if err != nil {
			return nil, isNewLead, err
		}
	} else if err != nil {
		return nil, isNewLead, err
	}

	return lead, isNewLead, nil
}
