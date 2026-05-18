package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"leadsextractor/messenger/flow"
	messengerstore "leadsextractor/messenger/store"
	"leadsextractor/messenger/handler"
	"leadsextractor/messenger/service"
	"leadsextractor/pkg/whatsapp"
	"leadsextractor/store"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	ctx := context.Background()

	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error cargando .env")
	}

	logger := slog.Default()

	db := store.ConnectDB(ctx)

	wpp := whatsapp.NewWhatsapp(
		os.Getenv("WHATSAPP_ACCESS_TOKEN"),
		os.Getenv("WHATSAPP_NUMBER_ID"),
		logger,
	)
	webhook := whatsapp.NewWebhook(os.Getenv("WHATSAPP_VERIFY_TOKEN"), logger)

	flowsPath := os.Getenv("MESSENGER_FLOWS_PATH")
	if flowsPath == "" {
		flowsPath = "../actions.json"
	}
	fm, err := flow.Load(flowsPath)
	if err != nil {
		log.Fatalf("Error cargando flows: %v", err)
	}

	msgStore := messengerstore.New(db)
	svc := service.New(msgStore, fm, logger)

	scheduler := service.NewScheduler(msgStore, wpp, 5*time.Second, logger)
	go scheduler.Start(ctx)

	h := handler.New(svc, webhook, logger)
	router := mux.NewRouter()
	h.RegisterRoutes(router)

	port := os.Getenv("MESSENGER_PORT")
	if port == "" {
		port = "8081"
	}
	addr := fmt.Sprintf(":%s", port)
	logger.Info("messenger iniciado", "addr", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
