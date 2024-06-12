package store

import (
	"fmt"
	"leadsextractor/models"
)

func (s *Store) GetAllAsesores(asesores *[]models.Asesor) error {
	if err := s.Db.Select(asesores, "SELECT * FROM Asesor"); err != nil {
		return err
	}
	return nil
}

func (s *Store) GetAllActiveAsesores(asesores *[]models.Asesor) error {
	if err := s.Db.Select(asesores, "SELECT * FROM Asesor where active=1"); err != nil {
		return err
	}
	return nil
}

func (s *Store) GetAllAsesoresExcept(phone string) (*[]models.Asesor, error) {
	asesores := []models.Asesor{}
	if err := s.Db.Select(&asesores, "SELECT * FROM Asesor where phone != ?", phone); err != nil {
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
	if err := s.Db.Select(&leads, query, phone); err != nil {
		return nil, err
	}
	return &leads, nil
}

func (s *Store) GetOneAsesor(phone string) (*models.Asesor, error) {
	asesor := models.Asesor{}
	if err := s.Db.Get(&asesor, "SELECT * FROM Asesor WHERE phone=?", phone); err != nil {
		return nil, err
	}
	return &asesor, nil
}

func (s *Store) GetAsesorFromEmail(email string) (*models.Asesor, error) {
	asesor := models.Asesor{}
	if err := s.Db.Get(&asesor, "SELECT * FROM Asesor WHERE email=?", email); err != nil {
		return nil, err
	}
	return &asesor, nil
}

func (s *Store) InsertAsesor(asesor *models.Asesor) (*models.Asesor, error) {
	query := "INSERT INTO Asesor (name, phone, active) VALUES (:name, :phone, :active)"
	if _, err := s.Db.NamedExec(query, asesor); err != nil {
		fmt.Printf("%v", err)
		return nil, err
	}
	return asesor, nil
}

func (s *Store) UpdateAsesor(asesor *models.Asesor, phone string) error {
	query := "UPDATE Asesor SET name=:name, active=:active WHERE phone=:phone"
	if _, err := s.Db.NamedExec(query, asesor); err != nil {
		return err
	}
	return nil
}
