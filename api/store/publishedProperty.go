package store

import (
	"fmt"
	"leadsextractor/models"

	"github.com/jmoiron/sqlx"
)

type PublishedStatus string

const (
    StatusNotPublished PublishedStatus = "not_published"
	StatusInQueue    PublishedStatus = "in_queue"
	StatusInProgress PublishedStatus = "in_progress"
	StatusPublished  PublishedStatus = "published"
	StatusFailed     PublishedStatus = "failed"
)

// Null fields because we do a left join
type PublishedProperty struct {
    PropertyID models.NullInt16  `json:"property_id" db:"property_id" validate:"required"`
    URL        models.NullString `json:"url"         db:"url"`
    Status     PublishedStatus   `json:"status"      db:"status" validate:"required"`
    Portal     models.NullString `json:"portal"      db:"portal" validate:"required"`
    // The internal ID of the property in one portal
    PublicationID models.NullString `json:"publication_id" db:"publication_id" validate:"required"`
    UpdatedAt  models.NullTime   `json:"updated_at"  db:"updated_at"`
    CreatedAt  models.NullTime   `json:"created_at"  db:"created_at"`
}

type UpdatePublishedProperty struct {
    Status          PublishedStatus `json:"status"`
    PublicationID   models.NullString  `json:"publication_id"`
}

type PublishedPropertyStorer interface {
	Create(pp *PublishedProperty) error
	GetOne(portal string, propertyID int64) (*PublishedProperty, error)
	GetAllByProp(propertyID int64) ([]*PublishedProperty, error)
	Update(portal string, propertyID int64, update *UpdatePublishedProperty) error
    GetAllByStatus(status PublishedStatus) ([]*PublishedProperty, error)
}

const (
    insertPublishedPropQ = `
		INSERT INTO PublishedProperty 
			(property_id, publication_id, status, portal) 
        VALUES (:property_id, :publication_id, :status, :portal)`
)

type publishedPropertyStore struct {
	db *sqlx.DB
}

func NewpublishedPropertyStore(db *sqlx.DB) PublishedPropertyStorer {
	return &publishedPropertyStore{db: db}
}

func (s *publishedPropertyStore) Create(pp *PublishedProperty) error {
	_, err := s.db.NamedExec(insertPublishedPropQ, pp)

    if err != nil {
        err = SQLBadForeignKey(err, 
            fmt.Sprintf("portal '%s' or propery with id %d does not exists", 
                pp.Portal.String, 
                pp.PropertyID.Int16,
            ))

        err = SQLDuplicated(err, fmt.Sprintf("property its already publicated in %s", pp.Portal.String))
    }
    return err
}

func (s *publishedPropertyStore) GetAllByProp(propertyID int64) ([]*PublishedProperty, error) {
	query := `
        SELECT 
            name as portal, 
            ifnull(status, "not_published") as status,
            property_id, publication_id, PP.url, updated_at, created_at
        FROM Portal P
        LEFT JOIN PublishedProperty PP
            ON P.name = PP.portal
            AND property_id = ?`

	var pp []*PublishedProperty

    err := s.db.Select(&pp, query, propertyID)
    if err != nil {
        return nil, fmt.Errorf("error obtaining publications by prop %w", err)
    }
    if len(pp) == 0 {
        panic("Are not any Portal in the db, insert at least one portal")
    }
	
	return pp, nil
}

func (s *publishedPropertyStore) GetAllByStatus(status PublishedStatus) ([]*PublishedProperty, error) {
	query := `
        SELECT 
           portal, status, property_id, publication_id, PP.url, updated_at, created_at
        FROM PublishedProperty PP
        WHERE status = ?`

	var pp []*PublishedProperty

    err := s.db.Select(&pp, query, status)
    if err != nil {
        return nil, fmt.Errorf("error obtaining publications by status %w", err)
    }
	
	return pp, nil
}

func (s *publishedPropertyStore) GetOne(portal string, propertyID int64) (*PublishedProperty, error) {
	query := `
		SELECT 
			property_id, publication_id, url, status, portal, updated_at, created_at 
		FROM PublishedProperty 
		WHERE portal = ? AND property_id = ?`

	var pp PublishedProperty
    err := s.db.Get(&pp, query, portal, propertyID)
    if err != nil {
        return nil, SQLNotFound(err, fmt.Sprintf("property (%d) is not published in %s", propertyID, portal)) 
    }
	
	return &pp, nil
}

func (s *publishedPropertyStore) Update(portal string, propertyID int64, update *UpdatePublishedProperty) error {
    setQuery := "SET status = :status"
    if update.PublicationID.Valid { // Only update the publication id if gives one valid
        setQuery += ", publication_id = :publication_id"
    }

    p := PublishedProperty{
        Status: update.Status,
        PublicationID: update.PublicationID,
        Portal: models.NullString{String: portal, Valid: true},
        PropertyID: models.NullInt16{Int16: int16(propertyID), Valid: true},
    }

	query := fmt.Sprintf(`
		UPDATE PublishedProperty %s 
        WHERE portal = :portal 
        AND property_id = :property_id
    `, setQuery)

	_, err := s.db.NamedExec(query, p)
	return SQLNotFound(err, "property does not exists")
}
