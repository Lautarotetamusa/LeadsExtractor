package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"leadsextractor/infobip"
	"leadsextractor/pipedrive"
	"leadsextractor/pkg"
	"leadsextractor/store"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

func main() {
    w := os.Stderr

    logger := slog.New(
        tint.NewHandler(w, &tint.Options{
            Level:      slog.LevelDebug,
            TimeFormat: time.DateTime,
        }),
    )

	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db := store.ConnectDB()

	apiPort := os.Getenv("API_PORT")
	host := fmt.Sprintf("%s:%s", "localhost", apiPort)

	infobipApi := infobip.NewInfobipApi(
        os.Getenv("INFOBIP_APIURL"),
        os.Getenv("INFOBIP_APIKEY"),
        "5213328092850",
        logger,
    )

	pipedriveApi := pipedrive.NewPipedrive(
		os.Getenv("PIPEDRIVE_CLIENT_ID"),
		os.Getenv("PIPEDRIVE_CLIENT_SECRET"),
		os.Getenv("PIPEDRIVE_API_TOKEN"),
		os.Getenv("PIPEDRIVE_REDIRECT_URI"),
        logger,
	)

	server := pkg.NewServer(host, logger, db, infobipApi, pipedriveApi)
	server.Run()
}
