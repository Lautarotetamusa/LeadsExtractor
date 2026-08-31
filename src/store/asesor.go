package store

import (
	"leadsextractor/models"

	"github.com/jmoiron/sqlx"
)

type AsesorStorer interface {
	GetAll() ([]*models.Asesor, error)
	GetAllActive() ([]*models.Asesor, error)
	GetLeads(phone string) ([]*models.Lead, error)
	GetOne(phone string) (*models.Asesor, error)
	GetFromEmail(email string) (*models.Asesor, error)
	Insert(*models.Asesor) error
	Update(*models.Asesor) error
	Delete(*models.Asesor) error
}

type AsesorDBStore struct {
	db *sqlx.DB
}

func NewAsesorDBStore(db *sqlx.DB) *AsesorDBStore {
	return &AsesorDBStore{
		db: db,
	}
}

func (s *AsesorDBStore) GetAll() ([]*models.Asesor, error) {
	var asesores []*models.Asesor
	if err := s.db.Select(&asesores, "SELECT * FROM Asesor"); err != nil {
		return nil, err
	}
	return asesores, nil
}

func (s *AsesorDBStore) GetAllActive() ([]*models.Asesor, error) {
	var asesores []*models.Asesor
	if err := s.db.Select(&asesores, "SELECT * FROM Asesor where active=1"); err != nil {
		return nil, err
	}
	return asesores, nil
}

func (s *AsesorDBStore) GetLeads(phone string) ([]*models.Lead, error) {
	query := `
        SELECT name, phone, email FROM Leads
        WHERE asesor=?
    `
	var leads []*models.Lead
	if err := s.db.Select(&leads, query, phone); err != nil {
		return nil, err
	}
	return leads, nil
}

func (s *AsesorDBStore) GetOne(phone string) (*models.Asesor, error) {
	var asesor models.Asesor
	if err := s.db.Get(&asesor, "SELECT * FROM Asesor WHERE phone=?", phone); err != nil {
		return nil, SQLNotFound(err, "asesor with this phone does not exists")
	}
	return &asesor, nil
}

func (s *AsesorDBStore) GetFromEmail(email string) (*models.Asesor, error) {
	var asesor models.Asesor
	if err := s.db.Get(&asesor, "SELECT * FROM Asesor WHERE email=?", email); err != nil {
		return nil, SQLNotFound(err, "asesor with this email does not exists")
	}
	return &asesor, nil
}

func (s *AsesorDBStore) Insert(asesor *models.Asesor) error {
	query := `
        INSERT INTO Asesor (name, phone, email, active) 
        VALUES (:name, :phone, :email, :active)`

	if _, err := s.db.NamedExec(query, asesor); err != nil {
		return SQLDuplicated(err, "already exists an asesor with this phone")
	}
	return nil
}

func (s *AsesorDBStore) Delete(a *models.Asesor) error {
	query := "DELETE FROM Asesor WHERE phone = :phone"
	if _, err := s.db.NamedExec(query, a); err != nil {
		return SQLNotFound(err, "asesor does not exists")
	}
	return nil
}

func (s *AsesorDBStore) Update(asesor *models.Asesor) error {
	query := `
    UPDATE Asesor 
    SET name=:name, 
        active=:active,
        email=:email,
        active=:active
    WHERE phone=:phone`

	if _, err := s.db.NamedExec(query, asesor); err != nil {
		return SQLNotFound(err, "asesor does not exists")
	}
	return nil
}
