package main

import (
	"context"
	"fmt"
	"leadsextractor/handlers"
	"leadsextractor/pkg/email"
	"leadsextractor/pkg/postmark"
	"leadsextractor/pkg/whatsapp"
	"leadsextractor/store"
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
	commStore := store.NewCommStore(db)

	useSMTP := false

	var mailer email.Sender

	if useSMTP {
		mailer = email.NewMailer(
			os.Getenv("SMTP_HOST"),
			os.Getenv("SMTP_PORT"),
			os.Getenv("SMTP_USER"),
			os.Getenv("SMTP_PASSWORD"),
			os.Getenv("SMTP_FROM"),
		)
	} else {
		mailer = postmark.NewClient(os.Getenv("POSTMARK_SERVER_TOKEN"), os.Getenv("POSTMARK_FROM"))
	}

	reportService := handlers.NewReportService(commStore, wa, mailer)

	reportNumbers := strings.Split(os.Getenv("REPORT_NUMBERS"), ";")
	fmt.Println(reportNumbers)

	// Send 99 days before for testing purposes
	if err := reportService.SendReport(reportNumbers, 999); err != nil {
		log.Fatalf("Error enviando reporte por WhatsApp: %v", err)
	}
	log.Println("Reporte WhatsApp enviado exitosamente")

	reportEmails := strings.Split(os.Getenv("REPORT_EMAILS"), ";")
	fmt.Println(reportEmails)

	if err := reportService.SendReportEmail(reportEmails, 999); err != nil {
		log.Fatalf("Error enviando reporte por email: %v", err)
	}
	log.Println("Reporte email enviado exitosamente")
}
