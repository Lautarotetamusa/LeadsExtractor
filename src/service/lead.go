package service

import (
	"leadsextractor/models"
	"leadsextractor/pkg/numbers"
	"leadsextractor/store"

	"github.com/go-playground/validator/v10"
)

type LeadService struct {
	lead     store.LeadStorer
	validate *validator.Validate
}

func NewLeadService(lead store.LeadStorer, validate *validator.Validate) *LeadService {
	return &LeadService{lead: lead, validate: validate}
}

func (s *LeadService) GetAll() (*[]models.Lead, error) {
	return s.lead.GetAll()
}

func (s *LeadService) GetOne(phone numbers.PhoneNumber) (*models.Lead, error) {
	return s.lead.GetOne(phone)
}

func (s *LeadService) Insert(createLead *models.CreateLead) (*models.Lead, error) {
	if err := s.validate.Struct(createLead); err != nil {
		return nil, NewValidationError(err.Error())
	}
	return s.lead.Insert(createLead)
}

func (s *LeadService) Update(phone numbers.PhoneNumber, update models.UpdateLead) (*models.Lead, error) {
	lead, err := s.lead.GetOne(phone)
	if err != nil {
		return nil, err
	}

	updateLeadsFields(lead, update)
	if err := s.validate.Struct(lead); err != nil {
		return nil, NewValidationError(err.Error())
	}

	if err := s.lead.Update(lead); err != nil {
		return nil, err
	}
	return lead, nil
}

func updateLeadsFields(lead *models.Lead, update models.UpdateLead) {
	if update.Name != "" {
		lead.Name = update.Name
	}
	if update.Email.Valid {
		lead.Email = update.Email
	}
}
