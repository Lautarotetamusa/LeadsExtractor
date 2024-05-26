package main

import (
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"leadsextractor/models"
	"leadsextractor/store"
	"leadsextractor/whatsapp"

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

    wpp := whatsapp.NewWhatsapp(
        os.Getenv("WHATSAPP_ACCESS_TOKEN"),
        os.Getenv("WHATSAPP_NUMBER_ID"),
        logger,
    )

    query := "CALL getCommunications(DATE_SUB(NOW(), interval 20 day))"

	communications := []models.Communication{}
	if err := db.Select(&communications, query); err != nil {
        log.Fatal(err)
	}

    const blockSize = 10
    var wg sync.WaitGroup

    for i := 0; i < len(communications); i += blockSize {
        end := i + blockSize
        if end > len(communications) {
            end = len(communications)
        }

        for _, c := range communications[i:end] {
            wg.Add(1)
            go func(c *models.Communication) {
                defer wg.Done()
                wpp.SendMsgAsesor(c.Asesor.Phone, c, true)
            }(&c)
        }

        wg.Wait()
    }

    wg.Wait()
}
