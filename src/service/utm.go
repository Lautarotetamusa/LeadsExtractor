package service

import (
	"fmt"
	"slices"
	"strings"
	"unicode"

	"leadsextractor/models"
	"leadsextractor/store"
)

type UTMService struct {
	utm store.UTMStorer
}

var validChannels = []string{"ivr", "whatsapp", "inbox"}

func NewUTMService(utm store.UTMStorer) *UTMService {
	return &UTMService{utm: utm}
}

func validateChannel(utm *models.UtmDefinition) error {
	if !slices.Contains(validChannels, utm.Channel.String) {
		return fmt.Errorf("the channel must be one of %s", strings.Join(validChannels, ", "))
	}
	return nil
}

func (s *UTMService) GetAll() ([]*models.UtmDefinition, error) {
	return s.utm.GetAll()
}

func (s *UTMService) GetOne(id int) (*models.UtmDefinition, error) {
	return s.utm.GetOne(id)
}

func (s *UTMService) Insert(utm *models.UtmDefinition) error {
	if err := validateChannel(utm); err != nil {
		return NewValidationError(err.Error())
	}

	if utm.Code == "" {
		return NewValidationError("code is required")
	}

	for _, r := range utm.Code {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return NewValidationError("code can only contains alphanumeric characters")
		}
	}
	utm.Code = strings.ToUpper(utm.Code)

	id, err := s.utm.Insert(utm)
	if err != nil {
		return err
	}
	utm.Id = int(id)
	return nil
}

// Update valida y persiste utm, que debe venir ya cargado (típicamente
// obtenido con GetOne y luego mutado in-place por el caller) para no perder
// los campos que no vinieron en un update parcial.
func (s *UTMService) Update(utm *models.UtmDefinition) error {
	if err := validateChannel(utm); err != nil {
		return NewValidationError(err.Error())
	}

	return s.utm.Update(utm)
}

func (s *UTMService) Delete(id int) error {
	return s.utm.Delete(id)
}
