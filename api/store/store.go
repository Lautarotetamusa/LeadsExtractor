package store

import (
	"leadsextractor/models"
	"leadsextractor/pkg/numbers"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type Storer interface {
    // Source storer
    GetSource(c *models.Communication) (*models.Source, error)
    insertSource(propertyId int) error

    // Property storer
    insertProperty(c *models.Communication, p *models.Propiedad) error
    insertOrGetProperty(c *models.Communication) (*models.Propiedad, error)

    // Message storer
    InsertMessage(m *Message) error

    // Action storer
    SaveAction(a *ActionSave) error
    GetLastActionFromLead(leadPhone numbers.PhoneNumber) (*ActionSave, error)
}

type Store struct {
	db *sqlx.DB
    logger *slog.Logger
}

func NewStore(db *sqlx.DB, logger *slog.Logger) *Store {
	return &Store{
		db: db,
        logger: logger.With("module", "store"),
	}
}
