package store

import (
	"fmt"
	"leadsextractor/models"
	"leadsextractor/numbers"
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

const (
    selectAllLeadQ = `SELECT 
        a.phone as "Asesor.phone", a.name as "Asesor.name",
        l.phone, l.name, l.email
        FROM Leads l
        INNER JOIN Asesor a
            ON a.phone = l.asesor`

    selectLeadQ = `SELECT 
        a.phone as "Asesor.phone", a.name as "Asesor.name", a.email as "Asesor.email",
        l.phone, l.name, l.email, l.cotizacion
        FROM Leads l
        INNER JOIN Asesor a
            ON a.phone = l.asesor
        WHERE l.phone=?`
)

func (s *LeadStore) GetAll() (*[]models.Lead, error) {
	leads := []models.Lead{}
	if err := s.db.Select(&leads, selectAllLeadQ); err != nil {
		return nil, err
	}
	return &leads, nil
}

func (s *LeadStore) GetOne(phone numbers.PhoneNumber) (*models.Lead, error) {
	lead := models.Lead{}
	if err := s.db.Get(&lead, selectLeadQ, phone); err != nil {
        return nil, SQLNotFound(err, fmt.Sprintf("the lead with phone %s does not exists", phone))
	}
	return &lead, nil
}

func (s *LeadStore) Insert(lead *models.CreateLead) (*models.Lead, error) {
	query := "INSERT INTO Leads (name, phone, email, asesor) VALUES (:name, :phone, :email, :asesor)"

	if _, err := s.db.NamedExec(query, lead); err != nil {
        return nil, SQLNotFound(err, fmt.Sprintf("the asesor with phone %s does not exists", lead.AsesorPhone))
	}

    // to get the id of the lead
	return s.GetOne(lead.Phone)
}

func (s *LeadStore) Update(lead *models.Lead, phone numbers.PhoneNumber) error {
	query := "UPDATE Leads SET name=:name, cotizacion=:cotizacion WHERE phone=:phone"

    _, err := s.db.NamedExec(query, lead);
    if err != nil {
        return SQLDuplicated(err, fmt.Sprintf("the lead with phone %s does not exists", phone))
	}

	return nil
}

func (s *LeadStore) UpdateAsesor(phone numbers.PhoneNumber, a *models.Asesor) error {
	query := "UPDATE Leads SET asesor=:asesor WHERE phone=:phone"
	_, err := s.db.NamedExec(query, map[string]interface{}{
        "asesor": a.Phone,
        "phone": phone.String(),
    });

    if err != nil {
        return SQLNotFound(err, fmt.Sprintf("the asesor with phone %s does not exists", a.Phone))
	}	

    return nil
}
