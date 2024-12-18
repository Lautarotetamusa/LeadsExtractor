package store

import (
	"leadsextractor/models"
)

func (s *Store) GetAllAsesores(asesores *[]models.Asesor) error {
	if err := s.db.Select(asesores, "SELECT * FROM Asesor"); err != nil {
		return err
	}
	return nil
}

func (s *Store) GetAllActiveAsesores(asesores *[]models.Asesor) error {
	if err := s.db.Select(asesores, "SELECT * FROM Asesor where active=1"); err != nil {
		return err
	}
	return nil
}

func (s *Store) GetAllAsesoresExcept(phone string) (*[]models.Asesor, error) {
	asesores := []models.Asesor{}
	if err := s.db.Select(&asesores, "SELECT * FROM Asesor where phone != ?", phone); err != nil {
		return nil, err
	}
	return &asesores, nil
}

func (s *Store) GetLeadsFromAsesor(phone string) (*[]models.Lead, error) {
	query := `
        SELECT name, phone, email FROM Leads
        WHERE asesor=?
    `
	leads := []models.Lead{}
	if err := s.db.Select(&leads, query, phone); err != nil {
		return nil, err
	}
	return &leads, nil
}

func (s *Store) GetOneAsesor(phone string) (*models.Asesor, error) {
	asesor := models.Asesor{}
	if err := s.db.Get(&asesor, "SELECT * FROM Asesor WHERE phone=?", phone); err != nil {
		return nil, err
	}
	return &asesor, nil
}

func (s *Store) GetAsesorFromEmail(email string) (*models.Asesor, error) {
	asesor := models.Asesor{}
	if err := s.db.Get(&asesor, "SELECT * FROM Asesor WHERE email=?", email); err != nil {
		return nil, err
	}
	return &asesor, nil
}

func (s *Store) InsertAsesor(asesor *models.Asesor) error {
	query := `INSERT INTO Asesor (name, phone, email, active) 
    VALUES (:name, :phone, :email, :active)`
	if _, err := s.db.NamedExec(query, asesor); err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteAsesor(a *models.Asesor) error {
    query := "DELETE FROM Asesor WHERE phone = :phone"
	if _, err := s.db.NamedExec(query, a); err != nil {
		return err
	}
	return nil
}

func (s *Store) UpdateAsesor(asesor *models.Asesor) error {
	query := `
    UPDATE Asesor 
    SET name=:name, 
        active=:active,
        email=:email,
        active=:active
    WHERE phone=:phone`

	if _, err := s.db.NamedExec(query, asesor); err != nil {
		return err
	}
	return nil
}
