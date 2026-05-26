package main

import (
	"context"
	"fmt"
	"leadsextractor/handlers"
	"leadsextractor/pkg/email"
	"leadsextractor/pkg/whatsapp"
	"leadsextractor/store"
	"leadsextractor/mocks"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	logger := slog.Default()

	db := store.ConnectDB(ctx)
	wa := whatsapp.NewWhatsapp(
		os.Getenv("WHATSAPP_ACCESS_TOKEN"),
		os.Getenv("WHATSAPP_NUMBER_ID"),
		logger,
	)

	mockStorer := true
	var commStore store.CommunicationStorer
	if mockStorer {
		commStore = &mocks.MockCommStorer{}
	} else {
		commStore = store.NewCommStore(db)
	}

	mailer := email.NewGraphMailer(email.Config{
		ClientID:		os.Getenv("MS_CLIENT_ID"),
		TenantID:		os.Getenv("MS_TENANT_ID"),
		ClientSecret: 	os.Getenv("MS_CLIENT_SECRET"),
		From:			"lautaro.teta@rbaresidences.com",
	})

	reportService := handlers.NewReportService(commStore, wa, mailer)

	reportNumbers := strings.Split(os.Getenv("REPORT_NUMBERS"), ";")
	fmt.Println(reportNumbers)

	// Send 99 days before for testing purposes
	if err := reportService.SendReport(reportNumbers, 99); err != nil {
		log.Fatalf("Error enviando reporte por WhatsApp: %v", err)
	}
	log.Println("Reporte WhatsApp enviado exitosamente")

	reportEmails := strings.Split(os.Getenv("REPORT_EMAILS"), ";")
	fmt.Println(reportEmails)

	if err := reportService.SendReportEmail(reportEmails, 99); err != nil {
		log.Fatalf("Error enviando reporte por email: %v", err)
	}
	log.Println("Reporte email enviado exitosamente")
}
