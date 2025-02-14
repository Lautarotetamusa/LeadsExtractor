package store

import (
	"leadsextractor/pkg/numbers"

	"github.com/jmoiron/sqlx"
)

// insertOrGetProperty(c *models.Communication) (*models.Propiedad, error)
// GetSource(c *models.Communication) (*models.Source, error)

type Storer interface {
    // Message storer
    InsertMessage(*Message) error

    // Action storer
    SaveAction(*ActionSave) error
    GetLastActionFromLead(numbers.PhoneNumber) (*ActionSave, error)
}

type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}
