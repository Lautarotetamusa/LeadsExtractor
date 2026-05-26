package main

import (
	"context"
	"flag"
	"fmt"
	"leadsextractor/handlers"
	"leadsextractor/mocks"
	"leadsextractor/pkg/email"
	"leadsextractor/pkg/whatsapp"
	"leadsextractor/store"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	days := flag.Int("days", 1, "cantidad de días previos a incluir en el reporte")
	useMock := flag.Bool("mock", false, "usar comunicaciones mockeadas en lugar de la base de datos")
	flag.Parse()

	ctx := context.Background()
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	logger := slog.Default()

	var commStore store.CommunicationStorer
	if *useMock {
		commStore = &mocks.MockCommStorer{}
	} else {
		db := store.ConnectDB(ctx)
		commStore = store.NewCommStore(db)
	}

	wa := whatsapp.NewWhatsapp(
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

	reportService := handlers.NewReportService(commStore, wa, mailer)

	reportNumbers := strings.Split(os.Getenv("REPORT_NUMBERS"), ";")
	fmt.Println(reportNumbers)

	if err := reportService.SendReport(reportNumbers, *days); err != nil {
		log.Fatalf("Error enviando reporte por WhatsApp: %v", err)
	}
	log.Println("Reporte WhatsApp enviado exitosamente")

	reportEmails := strings.Split(os.Getenv("REPORT_EMAILS"), ";")
	fmt.Println(reportEmails)

	if err := reportService.SendReportEmail(reportEmails, *days); err != nil {
		log.Fatalf("Error enviando reporte por email: %v", err)
	}
	log.Println("Reporte email enviado exitosamente")
}
