package service_test

import (
	"leadsextractor/mocks"
	"leadsextractor/models"
	"leadsextractor/service"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Antes de este refactor, validateChannel y el saneo de code eran funciones
// sueltas en el paquete handlers, solo probables vía HTTP. Acá se prueban
// directo sobre el service.
func TestUTMInsertValidation(t *testing.T) {
	utmStore := &mocks.MockUTMStorer{}
	utmStore.Mock()
	svc := service.NewUTMService(utmStore)

	t.Run("channel inválido", func(t *testing.T) {
		utm := &models.UtmDefinition{Code: "VALIDO", Channel: models.NullString{String: "invalido", Valid: true}}
		err := svc.Insert(utm)
		assert.Error(t, err)
		assert.IsType(t, service.ValidationError{}, err)
	})

	t.Run("code vacío", func(t *testing.T) {
		utm := &models.UtmDefinition{Code: "", Channel: models.NullString{String: "ivr", Valid: true}}
		err := svc.Insert(utm)
		assert.ErrorContains(t, err, "code is required")
	})

	t.Run("code no alfanumérico", func(t *testing.T) {
		utm := &models.UtmDefinition{Code: "__TEST__", Channel: models.NullString{String: "ivr", Valid: true}}
		err := svc.Insert(utm)
		assert.ErrorContains(t, err, "alphanumeric")
	})

	t.Run("code se normaliza a mayúsculas", func(t *testing.T) {
		utm := &models.UtmDefinition{Code: "minusculas", Channel: models.NullString{String: "ivr", Valid: true}}
		assert.NoError(t, svc.Insert(utm))
		assert.Equal(t, "MINUSCULAS", utm.Code)
	})
}
