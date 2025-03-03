package store

// property_id INT NOT NULL,
//
// id      VARCHAR(128) NOT NULL,
// url     VARCHAR(256) NOT NULL,
// status  ENUM("in_progress", "completed", "failed") DEFAULT "in_progress",
// portal  VARCHAR(64) NOT NULL,
//
// updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
// created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

import (
	"fmt"
	"leadsextractor/models"
	"time"

	"github.com/jmoiron/sqlx"
)

type PublishedStatus string

const (
	StatusInProgress PublishedStatus = "in_progress"
	StatusCompleted  PublishedStatus = "completed"
	StatusFailed     PublishedStatus = "failed"
)

type PublishedProperty struct {
    ID         int64            `json:"id" db:"id"`
    PropertyID int              `json:"property_id" validate:"required"`
	URL        models.NullString `json:"url"`
    Status     PublishedStatus  `json:"status" validate:"required"`
    Portal     string           `json:"portal" validate:"required"`
    UpdatedAt  time.Time        `json:"updated_at"`
	CreatedAt  time.Time        `json:"created_at"`
}

type PublishedPropertyStorer interface {
	Create(pp *PublishedProperty) error
	GetOne(portal string, propertyID int) (*PublishedProperty, error)
	UpdateStatus(portal string, propertyID int, status PublishedStatus) error
}

const (
    insertPublishedPropQ = `
		INSERT INTO PublishedProperty 
			(property_id, status, portal) 
        VALUES (:property_id, :status, :portal)`
)

type publishedPropertyStore struct {
	db *sqlx.DB
}

func NewpublishedPropertyStore(db *sqlx.DB) PublishedPropertyStorer {
	return &publishedPropertyStore{db: db}
}

func (s *publishedPropertyStore) Create(pp *PublishedProperty) error {
	res, err := s.db.NamedExec(insertPublishedPropQ, pp)
    pp.ID, err = res.LastInsertId()
    if err != nil {
        return fmt.Errorf("error getting the last inserted id %w", err)
    }

    return SQLBadForeignKey(err, fmt.Sprintf("portal '%s' or propery with id %d does not exists", pp.Portal, pp.PropertyID))
}

func (s *publishedPropertyStore) GetOne(portal string, propertyID int) (*PublishedProperty, error) {
	query := `
		SELECT 
			property_id, id, url, status, portal, updated_at, created_at 
		FROM PublishedProperty 
		WHERE portal = ? AND property_id = ?`

	var pp PublishedProperty
    err := s.db.Get(&pp, query, portal, propertyID)
	
	return &pp, SQLNotFound(err, fmt.Sprintf("property %d, does not is published in %s", propertyID, portal))
}

func (s *publishedPropertyStore) UpdateStatus(portal string, propertyID int, status PublishedStatus) error {
	query := `
		UPDATE PublishedProperty 
		SET status = ? 
		WHERE portal = ? AND property_id = ?`

	_, err := s.db.Exec(query, status, portal, propertyID)
	return SQLNotFound(err, "property does not exists")
}
