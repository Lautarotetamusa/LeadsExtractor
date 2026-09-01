package service_test

import (
	"leadsextractor/mocks"
	"leadsextractor/models"
	"leadsextractor/pkg/numbers"
	"leadsextractor/pkg/roundrobin"
	"leadsextractor/service"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestCommService() *service.CommunicationService {
	leadStore := &mocks.MockLeadStorer{}
	leadStore.Mock()

	asesorStore := mocks.NewMockAsesorStore()
	asesorStore.Mock()
	asesores, _ := asesorStore.GetAll()
	rr := roundrobin.New(asesores)

	return &service.CommunicationService{
		RoundRobin: rr,
		Logger:     slog.Default(),
		Leads:      leadStore,
		Utms:       &mocks.MockUTMStorer{},
		Properties: &mocks.MockPropertyStorer{},
	}
}

// Prueba de servicio: no pasa por HTTP, ni por un router, solo ejercita la
// lógica de negocio de dedup de leads con stores mockeados.
func TestSaveLead(t *testing.T) {
	commService := newTestCommService()

	c := &models.Communication{
		Fuente:    "inmuebles24",
		FechaLead: "2024-04-07",
		Fecha:     "2024-04-08",
		Nombre:    "Lautaro",
		Link:      "https://www.inmuebles24.com/panel/interesados/198059132",
		Telefono:  "5493415555555",
	}

	expected := &models.Lead{
		Name:   "Lautaro",
		Phone:  numbers.PhoneNumber("5493415555555"),
		Asesor: models.Asesor{Name: "test", Phone: "5493415556666", Email: "test@gmail.com", Active: false},
	}

	t.Run("new lead", func(t *testing.T) {
		lead, err := commService.SaveLead(c)
		assert.NoError(t, err)
		assert.Equal(t, expected, lead)
		assert.True(t, c.IsNew)
		assert.Equal(t, expected.Asesor.Phone, c.Asesor.Phone)
	})

	c = &models.Communication{
		Fuente:    "inmuebles24",
		FechaLead: "2024-04-07",
		Fecha:     "2024-04-08",
		Nombre:    "Lautaro",
		Link:      "https://www.inmuebles24.com/panel/interesados/198059132",
		Telefono:  "5493415555555",
	}

	t.Run("duplicated lead", func(t *testing.T) {
		lead, err := commService.SaveLead(c)
		assert.NoError(t, err)
		assert.Equal(t, expected, lead)
		assert.False(t, c.IsNew)
		assert.Equal(t, expected.Asesor.Phone, c.Asesor.Phone)
	})

	t.Run("duplicated lead with new data", func(t *testing.T) {
		// Now the lead have an email
		email := models.NullString{String: "cornejoy369@gmail.com", Valid: true}
		c = &models.Communication{
			Fuente:   "propiedades",
			Telefono: "5493415555555",
			Email:    email,
		}
		expected.Email = email

		lead, err := commService.SaveLead(c)
		assert.NoError(t, err)
		assert.Equal(t, expected, lead)
		assert.False(t, c.IsNew)
	})

	t.Run("duplicated lead with less data", func(t *testing.T) {
		c = &models.Communication{
			Fuente:   "ivr",
			Telefono: "5493415555555",
		}
		lead, err := commService.SaveLead(c)

		assert.NoError(t, err)
		assert.False(t, c.IsNew)
		assert.Equal(t, lead.Email, c.Email)
		assert.Equal(t, lead.Name, c.Nombre)
	})
}
