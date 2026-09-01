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
	"leadsextractor/pkg/email"
	"leadsextractor/pkg/infobip"
	"leadsextractor/pkg/pipedrive"
	"leadsextractor/pkg/roundrobin"
	"leadsextractor/pkg/whatsapp"
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
	if err := godotenv.Load("../.env"); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

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
