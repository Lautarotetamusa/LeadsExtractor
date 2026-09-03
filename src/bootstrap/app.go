// Package bootstrap arma el grafo de dependencias del proceso (DB, clientes
// externos, stores) en un único lugar, para que los distintos binarios
// (api, reporter) no dupliquen ni diverjan en cómo se construyen.
package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"leadsextractor/models"
	"leadsextractor/pkg/dependency"
	"leadsextractor/pkg/email"
	"leadsextractor/pkg/infobip"
	"leadsextractor/pkg/pipedrive"
	"leadsextractor/pkg/roundrobin"
	"leadsextractor/pkg/whatsapp"
	"leadsextractor/pkg/zenrows"
	"leadsextractor/store"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"

	_ "github.com/go-sql-driver/mysql"
)

type App struct {
	Logger *slog.Logger
	DB     *sqlx.DB

	Infobip   *infobip.InfobipApi
	Pipedrive *pipedrive.Pipedrive
	Whatsapp  *whatsapp.Whatsapp
	Mailer    email.Sender
	ZenRows   *zenrows.ZenRows

	LeadStore     store.LeadStorer
	UtmStore      store.UTMStorer
	CommStore     store.CommunicationStorer
	AsesorStore   store.AsesorStorer
	PropertyStore store.PropertyStorer
	Store         *store.Store

	RoundRobin *roundrobin.RoundRobin[models.Asesor]
	Validate   *validator.Validate
}

// NewApp carga el .env, conecta la DB, aplica las migraciones pendientes y
// construye todos los stores y clientes externos. Lo usan tanto el servicio
// HTTP principal como los binarios de cron (cmd/reporter) para no
// reconstruir a mano el mismo grafo de dependencias.
func NewApp(ctx context.Context) (*App, error) {
	// En Docker las env vars ya las inyecta `env_file:` de docker-compose, así
	// que no hay un .env físico dentro de la imagen (no se hornean secretos
	// en la imagen que se publica a GHCR). godotenv.Load solo hace falta
	// corriendo local sin Docker; si no encuentra el archivo, seguimos.
	_ = godotenv.Load("../.env")

	logger := slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.DateTime,
		}),
	)

	db := store.ConnectDB(ctx)
	if err := store.RunMigrations(db); err != nil {
		return nil, fmt.Errorf("error aplicando migraciones: %w", err)
	}

	infobipApi := infobip.NewInfobipApi(
		os.Getenv("INFOBIP_APIURL"),
		os.Getenv("INFOBIP_APIKEY"),
		"5213328092850",
		logger,
	)

	pipedriveApi := pipedrive.NewPipedrive(pipedrive.Config{
		ClientId:     os.Getenv("PIPEDRIVE_CLIENT_ID"),
		ClientSecret: os.Getenv("PIPEDRIVE_CLIENT_SECRET"),
		ApiToken:     os.Getenv("PIPEDRIVE_API_TOKEN"),
		RedirectURI:  os.Getenv("PIPEDRIVE_REDIRECT_URI"),
	}, logger)

	wpp := whatsapp.NewWhatsapp(
		os.Getenv("WHATSAPP_ACCESS_TOKEN"),
		os.Getenv("WHATSAPP_NUMBER_ID"),
		logger,
	)

	mailer := email.NewGraphMailer(email.Config{
		ClientID:     os.Getenv("MS_CLIENT_ID"),
		TenantID:     os.Getenv("MS_TENANT_ID"),
		ClientSecret: os.Getenv("MS_CLIENT_SECRET"),
		From:         "lautaro.teta@rbaresidences.com",
	})

	zenrowsApi := zenrows.NewZenRows(os.Getenv("ZENROWS_APIKEY"), logger)

	leadStore := store.NewLeadStore(db)
	utmStore := store.NewUTMStore(db)
	commStore := store.NewCommStore(db)
	asesorStore := store.NewAsesorDBStore(db)
	propertyStore := store.NewPropertyDBStore(db)
	storer := store.NewStore(db)

	asesores, err := asesorStore.GetAllActive()
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener la lista de asesores: %w", err)
	}
	rr := roundrobin.New(asesores)

	validate := validator.New(validator.WithRequiredStructEnabled())

	return &App{
		Logger: logger,
		DB:     db,

		Infobip:   infobipApi,
		Pipedrive: pipedriveApi,
		Whatsapp:  wpp,
		Mailer:    mailer,
		ZenRows:   zenrowsApi,

		LeadStore:     leadStore,
		UtmStore:      utmStore,
		CommStore:     commStore,
		AsesorStore:   asesorStore,
		PropertyStore: propertyStore,
		Store:         storer,

		RoundRobin: rr,
		Validate:   validate,
	}, nil
}

// Dependencies devuelve todas las APIs externas de las que depende la app
// (más la propia base de datos), para poder validarlas todas juntas con
// dependency.CheckAll.
func (a *App) Dependencies() []dependency.Dependency {
	return []dependency.Dependency{
		store.NewDBDependency(a.DB),
		a.Infobip,
		a.Pipedrive,
		a.Whatsapp,
		a.Mailer,
		a.ZenRows,
	}
}

// CheckRebora valida la salud de todo el ecosistema Rebora: las
// dependencias propias de este servicio, más Portalia (desglosado por cada
// portal que chequea internamente) y Cotizador. PORTALIA_URL/COTIZADOR_URL
// son las URLs base de cada servicio (sin path); si no están configuradas,
// esos chequeos se omiten en vez de reportarse como caídos.
func (a *App) CheckRebora(ctx context.Context) []dependency.Result {
	results := dependency.CheckAll(ctx, a.Dependencies(), 10*time.Second)

	if portaliaURL := os.Getenv("PORTALIA_URL"); portaliaURL != "" {
		results = append(results, dependency.FetchRemoteHealth(ctx, "portalia", portaliaURL+"/api/v1/health", 30*time.Second)...)
	}

	if cotizadorURL := os.Getenv("COTIZADOR_URL"); cotizadorURL != "" {
		results = append(results, dependency.CheckURL(ctx, "cotizador", cotizadorURL, 10*time.Second))
	}

	return results
}
