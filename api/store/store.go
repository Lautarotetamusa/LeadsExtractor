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

func (s *Store) InsertCotizacion(c *models.Communication) error {
    s.logger.Debug("Guardando cotizacion")
    query := `UPDATE Leads SET cotizacion = $1 WHERE phone = $2`
    if _, err := s.Db.Query(query, c.Cotizacion, c.Telefono); err != nil {
        return err
    }
    s.logger.Info("Cotizacion guardada con exito")
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
        if err := s.Update(lead, lead.Phone); err != nil {
            s.logger.Warn("error actualizando lead", "lead", lead.Phone)
        }

		return nil, err
	}

	return lead, nil
}
