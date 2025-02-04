package store

import (
	"fmt"
	"leadsextractor/models"
	"leadsextractor/numbers"
	"strings"
)

type LeadStorer interface {
    GetAll() (*[]models.Lead, error)
    GetOne(numbers.PhoneNumber) (*models.Lead, error)
    Insert(*models.CreateLead) (*models.Lead, error)
    Update(*models.Lead, numbers.PhoneNumber) error
    UpdateAsesor(numbers.PhoneNumber, *models.Asesor) error
}

type LeadStore struct {
    *Store
}

func (s *LeadStore) GetAll() (*[]models.Lead, error) {
	query := `SELECT 
        a.phone as "Asesor.phone", a.name as "Asesor.name",
        l.phone, l.name, l.email
        FROM Leads l
        INNER JOIN Asesor a
            ON a.phone = l.asesor`

	leads := []models.Lead{}
	if err := s.db.Select(&leads, query); err != nil {
		return nil, err
	}
	return &leads, nil
}

func (s *LeadStore) GetOne(phone numbers.PhoneNumber) (*models.Lead, error) {
	query := `SELECT 
        a.phone as "Asesor.phone", a.name as "Asesor.name", a.email as "Asesor.email",
        l.phone, l.name, l.email, l.cotizacion
        FROM Leads l
        INNER JOIN Asesor a
            ON a.phone = l.asesor
        WHERE l.phone=?`

	lead := models.Lead{}
	if err := s.db.Get(&lead, query, phone); err != nil {
		return nil, err
	}
	return &lead, nil
}

func (s *LeadStore) Insert(createLead *models.CreateLead) (*models.Lead, error) {
	query := "INSERT INTO Leads (name, phone, email, asesor) VALUES (:name, :phone, :email, :asesor)"
	if _, err := s.db.NamedExec(query, createLead); err != nil {
		if strings.Contains(err.Error(), "Error 1452") {
			return nil, fmt.Errorf("el asesor con telefono %s no existe", createLead.AsesorPhone)
		}
		return nil, err
	}
	lead, _ := s.GetOne(createLead.Phone)
	return lead, nil
}

func (s *LeadStore) Update(lead *models.Lead, phone numbers.PhoneNumber) error {
    s.logger.Debug("actualizando lead", "cotizacion", lead.Cotizacion)
	query := "UPDATE Leads SET name=:name, cotizacion=:cotizacion WHERE phone=:phone"
	res, err := s.db.NamedExec(query, lead);

    if err != nil {
        s.logger.Warn("error actualizando lead", "lead", lead.Phone)
		return err
	}

    afid, _ := res.RowsAffected()
    s.logger.Info("lead actualizado", "rows", afid)
	return nil
}

func (s *LeadStore) UpdateAsesor(phone numbers.PhoneNumber, a *models.Asesor) error {
	query := "UPDATE Leads SET asesor=:asesor WHERE phone=:phone"
	_, err := s.db.NamedExec(query, map[string]interface{}{
        "asesor": a.Phone,
        "phone": phone.String(),
    });

    if err != nil {
		s.logger.Error(fmt.Sprintf("%v\n", err))
		return err
	}	

    return nil
}
