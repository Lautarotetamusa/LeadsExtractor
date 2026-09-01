package service

import (
	"leadsextractor/mocks"
	"leadsextractor/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Antes de mover CommunicationService a este paquete, probar esto exigía
// levantar el router HTTP entero (handlers_test.TestMain). Al ser un test
// interno del paquete service, alcanza con mockear las stores.
func TestAttachProperty(t *testing.T) {
	commService := &CommunicationService{
		Properties: &mocks.MockPropertyStorer{},
	}

	t.Run("whatsapp/ivr/viewphone no tienen property asociada", func(t *testing.T) {
		for _, fuente := range []string{"whatsapp", "ivr", "viewphone"} {
			c := &models.Communication{Fuente: fuente}
			err := commService.attachProperty(c)
			assert.NoError(t, err)
			assert.False(t, c.Propiedad.ID.Valid)
		}
	})

	t.Run("fuente inválida", func(t *testing.T) {
		c := &models.Communication{Fuente: "portal_inventado"}
		err := commService.attachProperty(c)
		assert.Error(t, err)
	})

	t.Run("crea la property si no existe y la reusa la segunda vez", func(t *testing.T) {
		c1 := &models.Communication{Fuente: "inmuebles24"}
		c1.Propiedad.PortalId = models.NullString{String: "ABC123", Valid: true}
		assert.NoError(t, commService.attachProperty(c1))
		assert.True(t, c1.Propiedad.ID.Valid)

		c2 := &models.Communication{Fuente: "inmuebles24"}
		c2.Propiedad.PortalId = models.NullString{String: "ABC123", Valid: true}
		assert.NoError(t, commService.attachProperty(c2))
		assert.Equal(t, c1.Propiedad.ID, c2.Propiedad.ID)
	})
}
