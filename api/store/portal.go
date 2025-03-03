package store

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Portal struct {
    Name    string  `json:"name" db:"name"`
    Url     string  `json:"url" db:"url"`
}

type PortalStorer interface {
    GetAll() ([]*Portal, error)
    GetOne(name string) (*Portal, error)
    Insert(*Portal) error
}

type portalDBStore struct {
	db *sqlx.DB
}

func NewPortalDBStore(db *sqlx.DB) PortalStorer {
	return &portalDBStore{db: db}
}

func (s *portalDBStore) GetAll() ([]*Portal, error) {
    portals := make([]*Portal, 0)

    query := "SELECT * FROM Portal"
	err := s.db.Select(&portals, query)
	if err != nil {
		return nil, fmt.Errorf("error obtaining all the portals: %w", err)
	}

	return portals, nil
}

func (s *portalDBStore) GetOne(name string) (*Portal, error) {
    var portal Portal

    query := "SELECT * FROM Portal WHERE name = ?"
	err := s.db.Get(&portal, query, name)
	if err != nil {
		return nil, SQLNotFound(err, fmt.Sprintf("portal with name %s does not exists", name))
	}

	return &portal, nil
}

func (s *portalDBStore) Insert(p *Portal) error {
    query := "INSERT INTO Portal (name, url) VALUES (:name, :url)"
	_, err := s.db.NamedExec(query, p)
	if err != nil {
        return SQLDuplicated(err, fmt.Sprintf("already exists a portal with name %s", p.Name))
	}

    return nil
}
