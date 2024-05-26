package store

import (
	"fmt"
	"leadsextractor/models"
	"strings"
)

func (s *Store) GetAll() (*[]models.Lead, error) {
	query := `SELECT 
        a.phone as "Asesor.phone", a.name as "Asesor.name",
        l.phone, l.name, l.email
        FROM Leads l
        INNER JOIN Asesor a
            ON a.phone = l.asesor`

	leads := []models.Lead{}
	if err := s.Db.Select(&leads, query); err != nil {
		return nil, err
	}
	return &leads, nil
}

func (s *Store) GetOne(phone string) (*models.Lead, error) {
	query := `SELECT 
        a.phone as "Asesor.phone", a.name as "Asesor.name", a.email as "Asesor.email",
        l.phone, l.name, l.email
        FROM Leads l
        INNER JOIN Asesor a
            ON a.phone = l.asesor
        WHERE l.phone=?`

	lead := models.Lead{}
	if err := s.Db.Get(&lead, query, phone); err != nil {
		return nil, err
	}
	return &lead, nil
}

func (s *Store) Insert(createLead *models.CreateLead) (*models.Lead, error) {
	query := "INSERT INTO Leads (name, phone, email, asesor) VALUES (:name, :phone, :email, :asesor)"
	if _, err := s.Db.NamedExec(query, createLead); err != nil {
		if strings.Contains(err.Error(), "Error 1452") {
			return nil, fmt.Errorf("el asesor con telefono %s no existe", createLead.AsesorPhone)
		}
		return nil, err
	}
	lead, _ := s.GetOne(createLead.Phone)
	return lead, nil
}

func (s *Store) Update(lead *models.Lead, phone string) (*models.Lead, error) {
	query := "UPDATE Leads SET name=:name, active=:active WHERE phone=:phone"
	if _, err := s.Db.NamedExec(query, lead); err != nil {
		fmt.Printf("%v", err)
		return nil, err
	}
	return lead, nil
}
