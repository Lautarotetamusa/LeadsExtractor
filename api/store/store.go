package store

import (
	"database/sql"
	"leadsextractor/models"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type Store struct {
	Db *sqlx.DB
    logger *slog.Logger
}

func NewStore(db *sqlx.DB, logger *slog.Logger) *Store {
	return &Store{
		Db: db,
        logger: logger.With("module", "store"),
	}
}

func (s *Store) InsertCommunication(c *models.Communication, lead *models.Lead, source *models.Source) error {
    s.logger.Debug("Saving communication")
	query := `INSERT INTO Communication(lead_phone, source_id, new_lead, lead_date, url, zones, mt2_terrain, mt2_builded, baths, rooms) 
    VALUES (:lead_phone, :source_id, :new_lead, :lead_date, :url, :zones, :mt2_terrain, :mt2_builded, :baths, :rooms)`
    _, err := s.Db.NamedExec(query, map[string]interface{}{
		"lead_phone":  lead.Phone,
		"source_id":   source.Id,
		"new_lead":    c.IsNew,
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
    s.logger.Info("communication saved")
    return nil
}

func (s *Store) InsertOrGetLead(rr *RoundRobin, c *models.Communication) (*models.Lead, error) {
	var lead *models.Lead

	lead, err := s.GetOne(c.Telefono)

	if err == sql.ErrNoRows {
		c.IsNew = true
		c.Asesor = rr.Next()

		lead, err = s.Insert(&models.CreateLead{
			Name:        c.Nombre,
			Phone:       c.Telefono,
			Email:       c.Email,
			AsesorPhone: c.Asesor.Phone,
            Cotizacion:  c.Cotizacion,
		})

		if err != nil {
			return nil, err
		}
	} else if err != nil {
        lead.Cotizacion = c.Cotizacion
        if err := s.Update(lead, lead.Phone); err != nil {
            s.logger.Warn("error actualizando lead", "lead", lead.Phone)
        }

		return nil, err
	}

	return lead, nil
}
