package service_test

import (
	"leadsextractor/mocks"
	"leadsextractor/models"
	"leadsextractor/pkg/numbers"
	"leadsextractor/pkg/roundrobin"
	"leadsextractor/service"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

// Antes de este refactor, ReasignLeads solo se podía probar disparando un
// PUT /asesor/{phone}/reasign contra el router HTTP completo. Acá se prueba
// directo, sin HTTP, con las stores mockeadas.
func TestReasignLeads(t *testing.T) {
	asesorStore := mocks.NewMockAsesorStore()
	asesorStore.Mock()
	leadStore := &mocks.MockLeadStorer{}
	leadStore.Mock()

	// MockAsesorStorer y MockLeadStorer no comparten datos: para que
	// UpdateAsesor encuentre los leads del asesor hay que sembrarlos acá también.
	for _, phone := range []string{"5493414449999", "5493414448888", "5493414447777"} {
		_, err := leadStore.Insert(&models.CreateLead{
			Name:        "lead",
			Phone:       numbers.PhoneNumber(phone),
			AsesorPhone: numbers.PhoneNumber("5493415554444"),
		})
		assert.NoError(t, err)
	}

	asesores, _ := asesorStore.GetAll()
	rr := roundrobin.New(asesores)
	validate := validator.New(validator.WithRequiredStructEnabled())

	svc := service.NewAsesorService(asesorStore, leadStore, rr, validate)

	n, err := svc.ReasignLeads("5493415554444")
	assert.NoError(t, err)
	assert.Equal(t, 3, n) // mocks.MockAsesorStorer siembra 3 leads para este asesor

	asesor, err := svc.GetOne("5493415554444")
	assert.NoError(t, err)
	assert.False(t, asesor.Active, "el asesor reasignado debe quedar inactivo")
}

func TestReasignLeadsAsesorNoExiste(t *testing.T) {
	asesorStore := mocks.NewMockAsesorStore()
	asesorStore.Mock()
	leadStore := &mocks.MockLeadStorer{}
	leadStore.Mock()

	asesores, _ := asesorStore.GetAll()
	rr := roundrobin.New(asesores)
	validate := validator.New(validator.WithRequiredStructEnabled())

	svc := service.NewAsesorService(asesorStore, leadStore, rr, validate)

	_, err := svc.ReasignLeads("no-existe")
	assert.Error(t, err)
}
