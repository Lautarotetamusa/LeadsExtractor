package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"leadsextractor/bootstrap"
	"leadsextractor/flow"
	"leadsextractor/handlers"
	"leadsextractor/pkg"
	"leadsextractor/pkg/whatsapp"
	"leadsextractor/service"

	"github.com/gorilla/mux"
)

func main() {
	ctx := context.Background()

	app, err := bootstrap.NewApp(ctx)
	if err != nil {
		log.Fatal(err)
	}

	webhook := whatsapp.NewWebhook(
		os.Getenv("WHATSAPP_VERIFY_TOKEN"),
		app.Logger,
	)

	healthCheckNumbers := strings.Split(os.Getenv("REPORT_NUMBERS"), ";")

	flowManager := flow.NewFlowManager("actions.json", app.Store, app.Logger)
	flow.DefineActions(app.Whatsapp, app.Pipedrive, app.Infobip, app.LeadStore, app.Mailer, app.Dependencies(), healthCheckNumbers)
	flowManager.MustLoad()

	// Services
	commsService := service.CommunicationService{
		RoundRobin: app.RoundRobin,
		Logger:     app.Logger,
		Flows:      *flowManager,
		Store:      app.Store,

		Properties: app.PropertyStore,
		Utms:       app.UtmStore,
		Comms:      app.CommStore,
		Leads:      app.LeadStore,
	}
	asesorService := service.NewAsesorService(app.AsesorStore, app.LeadStore, app.RoundRobin, app.Validate)
	utmService := service.NewUTMService(app.UtmStore)
	leadService := service.NewLeadService(app.LeadStore, app.Validate)

	// Handlers
	leadHandler := handlers.NewLeadHandler(leadService)
	utmHandler := handlers.NewUTMHandler(utmService)
	flowHandler := handlers.NewFlowHandler(flowManager, app.CommStore)
	commHandler := handlers.NewCommHandler(commsService)
	asesorHandler := handlers.NewAsesorHandler(asesorService)
	healthHandler := handlers.NewHealthHandler(app.Dependencies())

	router := mux.NewRouter()
	router.Use(CORS)

	// Register routes
	leadHandler.RegisterRoutes(router)
	utmHandler.RegisterRoutes(router)
	flowHandler.RegisterRoutes(router)
	commHandler.RegisterRoutes(router)
	asesorHandler.RegisterRoutes(router)
	healthHandler.RegisterRoutes(router)

	// Server
	apiPort := os.Getenv("API_PORT")
	host := fmt.Sprintf("%s:%s", "localhost", apiPort)
	server := pkg.NewServer(pkg.ServerOpts{
		ListenAddr: host,
		Logger:     app.Logger,
	})

	go webhook.ConsumeEntries(commsService.NewCommunication)

	aircall := pkg.NewAircall(commsService.NewCommunication, app.Logger)
	router.Handle("/aircall", aircall).Methods(http.MethodPost)

	router.HandleFunc("/pipedrive", handlers.HandleErrors(app.Pipedrive.HandleOAuth)).Methods(http.MethodGet)

	router.HandleFunc("/webhooks", handlers.HandleErrors(webhook.ReciveNotificaction)).Methods(http.MethodPost)
	router.HandleFunc("/webhooks", handlers.HandleErrors(webhook.Verify)).Methods(http.MethodGet)

	server.Run(router)
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, DELETE, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
