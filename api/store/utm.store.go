package store

import (
	"leadsextractor/models"

	"github.com/jmoiron/sqlx"
)

type UTMStorer interface {
    GetAll() ([]*models.UtmDefinition, error)
    GetOne(int) (*models.UtmDefinition, error)
    GetOneByCode(string) (*models.UtmDefinition, error)
    Insert(*models.UtmDefinition) (int64, error)
    Update(*models.UtmDefinition) error
    Delete(int) error
}

type UTMStore struct {
    db *sqlx.DB
}

func NewUTMStore(db *sqlx.DB) *UTMStore {
    return &UTMStore{
        db: db,
    }
}

const (
    updateUTMQuery = `
        UPDATE Utm 
        SET utm_source = :utm_source,
            utm_medium = :utm_medium ,
            utm_campaign = :utm_campaign,
            utm_ad = :utm_ad,
            utm_channel = :utm_channel
        WHERE id=:id`

    insertUTMQuery = `
    INSERT INTO Utm 
            ( code,  utm_source, utm_medium,  utm_campaign,  utm_ad,  utm_channel) 
    VALUES  (:code, :utm_source, :utm_medium, :utm_campaign, :utm_ad, :utm_channel)`
)

func (s *UTMStore) GetAll() ([]*models.UtmDefinition, error) {
    var utms []*models.UtmDefinition
	if err := s.db.Select(&utms, "SELECT * FROM Utm ORDER BY id DESC"); err != nil {
		return nil, err
	}
	return utms, nil
}

func (s *UTMStore) GetOne(id int) (*models.UtmDefinition, error) {
	utm := models.UtmDefinition{}
	if err := s.db.Get(&utm, "SELECT * FROM Utm WHERE id=?", id); err != nil {
		return nil, SQLNotFound(err, "utm not found")
	}
	return &utm, nil
}

func (s *UTMStore) GetOneByCode(code string) (*models.UtmDefinition, error) {
	utm := models.UtmDefinition{}
	if err := s.db.Get(&utm, "SELECT * FROM Utm WHERE code=?", code); err != nil {
		return nil, SQLNotFound(err, "utm not found")
	}
	return &utm, nil
}

func (s *UTMStore) Insert(utm *models.UtmDefinition) (int64, error) {
    res, err := s.db.NamedExec(insertUTMQuery, utm)
    if err != nil {
        return 0, SQLDuplicated(err, "utm with this code already exists")
    }

    id, err := res.LastInsertId()
    if err != nil {
        return 0, err
    }
	return id, nil
}

func (s *UTMStore) Update(utm *models.UtmDefinition) error {
	if _, err := s.db.NamedExec(updateUTMQuery, utm); err != nil {
        return SQLDuplicated(err, "utm with this code already exists")
	}
	return nil
}

func (s *UTMStore) Delete(id int) error {
    query := "DELETE FROM Utm where id=?"

	if _, err := s.db.Exec(query, id); err != nil {
		return SQLNotFound(err, "utm not found")
	}
	return nil
}
