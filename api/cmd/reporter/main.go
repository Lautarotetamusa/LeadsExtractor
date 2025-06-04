package main

import (
	"context"
	"fmt"
	"leadsextractor/handlers"
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
    
    reportService := handlers.NewReportService(commStore, wa)
    
    reportNumbers := strings.Split(os.Getenv("REPORT_NUMBERS"), ";")
    fmt.Println(reportNumbers)
    
    if err := reportService.SendDailyReport(reportNumbers); err != nil {
        log.Fatalf("Error enviando reporte: %v", err)
    }
    
    log.Println("Reporte enviado exitosamente")
}
