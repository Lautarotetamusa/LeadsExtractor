package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"leadsextractor/bootstrap"
	"leadsextractor/mocks"
	"leadsextractor/pkg/email"
	"leadsextractor/pkg/whatsapp"
	"leadsextractor/service"
	"leadsextractor/store"

	"github.com/joho/godotenv"
)

func main() {
	days := flag.Int("days", 1, "cantidad de días previos a incluir en el reporte")
	useMock := flag.Bool("mock", false, "usar comunicaciones mockeadas en lugar de la base de datos")
	flag.Parse()

	ctx := context.Background()

	var commStore store.CommunicationStorer
	var wa *whatsapp.Whatsapp
	var mailer email.Sender

	if *useMock {
		if err := godotenv.Load("../.env"); err != nil {
			log.Fatal("Error loading .env file")
		}
		commStore = &mocks.MockCommStorer{}
		wa = whatsapp.NewWhatsapp(
			os.Getenv("WHATSAPP_ACCESS_TOKEN"),
			os.Getenv("WHATSAPP_NUMBER_ID"),
			slog.Default(),
		)
		mailer = email.NewGraphMailer(email.Config{
			ClientID:     os.Getenv("MS_CLIENT_ID"),
			TenantID:     os.Getenv("MS_TENANT_ID"),
			ClientSecret: os.Getenv("MS_CLIENT_SECRET"),
			From:         "lautaro.teta@rbaresidences.com",
		})
	} else {
		app, err := bootstrap.NewApp(ctx)
		if err != nil {
			log.Fatal(err)
		}
		commStore = app.CommStore
		wa = app.Whatsapp
		mailer = app.Mailer
	}

	reportService := service.NewReportService(commStore, wa, mailer)

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
