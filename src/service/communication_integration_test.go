//go:build integration

// Tests de integración: a diferencia del resto de la suite (mockeada), estos
// corren migraciones reales contra una base descartable y ejercitan
// StoreCommunication con stores reales — es la única forma de detectar un
// desajuste entre el código y el schema real (una columna que el INSERT
// espera pero la migración nunca creó, un valor de ENUM que falta, etc.).
// Un mock nunca puede fallar así, porque nunca corre SQL de verdad; ver
// HEALTHCHECK.md y el commit que agrega la migración 00004 para el caso
// real que motivó esto (bedrooms/bathrooms/total_area/covered_area nunca
// existieron en producción pese a que goose las marcaba como aplicadas).
//
// Correr con: go test -tags=integration ./service/...
// Requiere las mismas env vars que bootstrap.NewApp (HOST, DB_USER,
// DB_PORT, DB_PASS, DB_NAME) apuntando a una DB descartable — NUNCA a
// producción, esto inserta filas de verdad.
package service_test

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"testing"
	"time"

	"leadsextractor/models"
	"leadsextractor/pkg/numbers"
	"leadsextractor/pkg/roundrobin"
	"leadsextractor/service"
	"leadsextractor/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newIntegrationCommService(t *testing.T) *service.CommunicationService {
	t.Helper()

	db := store.ConnectDB(context.Background())
	require.NoError(t, store.RunMigrations(db))

	asesores, err := store.NewAsesorDBStore(db).GetAllActive()
	require.NoError(t, err)
	require.NotEmpty(t, asesores, "necesita al menos un asesor activo (lo siembra 00001_baseline.sql)")

	return &service.CommunicationService{
		RoundRobin: roundrobin.New(asesores),
		Logger:     slog.Default(),
		Leads:      store.NewLeadStore(db),
		Utms:       store.NewUTMStore(db),
		Comms:      store.NewCommStore(db),
		Properties: store.NewPropertyDBStore(db),
		Store:      store.NewStore(db),
	}
}

// TestStoreCommunication_AllSources ejercita attachProperty + SaveLead +
// Insert (todo el camino de StoreCommunication) contra una DB real, para
// cada una de las fuentes válidas — incluyendo las que traen Property
// (donde vivía el bug real: InsertProperty solo se llama para properties
// nuevas, así que un desajuste de schema recién se ve acá, no en el GET).
func TestStoreCommunication_AllSources(t *testing.T) {
	commService := newIntegrationCommService(t)
	run := rand.New(rand.NewSource(time.Now().UnixNano())).Int63() % 1_000_000

	for i, fuente := range store.ValidSources() {
		fuente := fuente
		phone := numbers.PhoneNumber(fmt.Sprintf("549341%07d", (run+int64(i))%10_000_000))

		t.Run(fuente, func(t *testing.T) {
			c := &models.Communication{
				Fuente:    fuente,
				FechaLead: time.Now().Format("2006-01-02"),
				Fecha:     time.Now().Format("2006-01-02"),
				Nombre:    "Test Integration",
				Telefono:  phone,
			}

			if store.FuenteTienePropiedad(fuente) {
				c.Propiedad = models.Propiedad{
					PortalId:    models.NullString{String: fmt.Sprintf("test-%d-%s", run, fuente), Valid: true},
					Titulo:      models.NullString{String: "Depto de prueba", Valid: true},
					Link:        models.NullString{String: "https://example.com/test", Valid: true},
					Precio:      models.NullString{String: "1000000", Valid: true},
					Ubicacion:   models.NullString{String: "Guadalajara", Valid: true},
					Bedrooms:    "3",
					Bathrooms:   "2",
					TotalArea:   "120",
					CoveredArea: "100",
				}
			}

			err := commService.StoreCommunication(c)
			assert.NoError(t, err)
		})
	}
}

// TestStoreCommunication_InvalidSource confirma que una fuente no
// reconocida se rechaza sin tocar la DB (no debería ni intentar el INSERT).
func TestStoreCommunication_InvalidSource(t *testing.T) {
	commService := newIntegrationCommService(t)

	c := &models.Communication{
		Fuente:    "fuente-que-no-existe",
		FechaLead: time.Now().Format("2006-01-02"),
		Telefono:  numbers.PhoneNumber("5493410000099"),
	}

	err := commService.StoreCommunication(c)
	assert.Error(t, err)
}
