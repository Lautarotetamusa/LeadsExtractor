package pipedrive

import (
	"fmt"
	"leadsextractor/models"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

func TestCreateDeal(t *testing.T) {
	c := &models.Communication{
		Fuente:    "inmuebles24",
		FechaLead: "2024-04-07",
		LeadId:    "461161340",
		Fecha:     "2024-04-08",
		Nombre:    "Lautaro",
		Link:      "https://www.inmuebles24.com/panel/interesados/198059132",
		Telefono:  "5493415854220",
        Email:     models.NullString{String: "cornejoy369@gmail.com", Valid: true},
		Propiedad: models.Propiedad{},
	}

    w := os.Stderr
    logger := slog.New(
        tint.NewHandler(w, &tint.Options{
            Level:      slog.LevelDebug,
            TimeFormat: time.DateTime,
        }),
    )

	if err := godotenv.Load("../../.env"); err != nil {
		logger.Error("Error loading .env file")
        os.Exit(1)
	}

    pipedriveConfig := Config{
        ClientId: os.Getenv("PIPEDRIVE_CLIENT_ID"),
        ClientSecret: os.Getenv("PIPEDRIVE_CLIENT_SECRET"),
        ApiToken: os.Getenv("PIPEDRIVE_API_TOKEN"),
        RedirectURI: os.Getenv("PIPEDRIVE_REDIRECT_URI"),
        FilePath: "../",
    }
	api := NewPipedrive(pipedriveConfig, logger)

    deal, err := api.createDeal(c, 13467041, 2426)
    if err != nil {
        t.Error(err.Error())
    }
    fmt.Printf("%v\n", deal)
}
