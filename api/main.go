package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"time"

	"leadsextractor/flow"
	"leadsextractor/handlers"
	"leadsextractor/pkg"
	"leadsextractor/pkg/infobip"
	"leadsextractor/pkg/pipedrive"
	"leadsextractor/pkg/roundrobin"
	"leadsextractor/pkg/whatsapp"
	"leadsextractor/store"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/lmittmann/tint"

	"github.com/robfig/cron/v3"
)

func main() {
	ctx := context.Background()

	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	logger := slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.DateTime,
		}),
	)

	db := store.ConnectDB(ctx)

    appHost := os.Getenv("APP_HOST")

	infobipApi := infobip.NewInfobipApi(
		os.Getenv("INFOBIP_APIURL"),
		os.Getenv("INFOBIP_APIKEY"),
		"5213328092850",
		logger,
	)

	pipedriveConfig := pipedrive.Config{
		ClientId:     os.Getenv("PIPEDRIVE_CLIENT_ID"),
		ClientSecret: os.Getenv("PIPEDRIVE_CLIENT_SECRET"),
		ApiToken:     os.Getenv("PIPEDRIVE_API_TOKEN"),
		RedirectURI:  os.Getenv("PIPEDRIVE_REDIRECT_URI"),
	}
	pipedriveApi := pipedrive.NewPipedrive(pipedriveConfig, logger)

	wpp := whatsapp.NewWhatsapp(
		os.Getenv("WHATSAPP_ACCESS_TOKEN"),
		os.Getenv("WHATSAPP_NUMBER_ID"),
		logger,
	)

	webhook := whatsapp.NewWebhook(
		os.Getenv("WHATSAPP_VERIFY_TOKEN"),
		logger,
	)

	// Stores
	storer := store.NewStore(db)

    propStore := store.NewPropertyPortalStore(db)
    publisPropStore := store.NewpublishedPropertyStore(db)
	leadStore := store.NewLeadStore(db)
	utmStore := store.NewUTMStore(db)
	commStore := store.NewCommStore(db)
	asesorStore := store.NewAsesorDBStore(db)
	sourceStore := store.NewSourceDBStore(db)

	// Round Robin
	asesores, err := asesorStore.GetAllActive()
	if err != nil {
		log.Fatal("No se pudo obtener la lista de asesores\nERROR: ", err.Error())
	}
	rr := roundrobin.New(asesores)

	flowManager := flow.NewFlowManager("actions.json", storer, logger)
	flow.DefineActions(wpp, pipedriveApi, infobipApi, leadStore)
	flowManager.MustLoad()

	// Services
	commsService := handlers.CommunicationService{
		RoundRobin: rr,
		Logger:     logger,
		Flows:      *flowManager,
		Store:      storer,

		Source: sourceStore,
		Utms:   utmStore,
		Comms:  commStore,
		Leads:  leadStore,
	}
	asesorService := handlers.NewAsesorService(asesorStore, leadStore, rr)

	// Handlers
    propHandler := handlers.NewPropertyHandler(propStore, logger)
    publishPropHandler := handlers.NewPublishedPropertyHandler(publisPropStore, propStore, appHost)
	leadHandler := handlers.NewLeadHandler(leadStore)
	utmHandler := handlers.NewUTMHandler(utmStore)
	flowHandler := handlers.NewFlowHandler(flowManager, commStore)
	commHandler := handlers.NewCommHandler(commsService)
	asesorHandler := handlers.NewAsesorHandler(asesorService)

	router := mux.NewRouter()
    router.Use(CORS)

	// Register routes
    propHandler.RegisterRoutes(router)
	leadHandler.RegisterRoutes(router)
	utmHandler.RegisterRoutes(router)
	flowHandler.RegisterRoutes(router)
	commHandler.RegisterRoutes(router)
	asesorHandler.RegisterRoutes(router)
    publishPropHandler.RegisterRoutes(router) 

    // Cron report
    reportService := handlers.NewReportService(commStore, wpp)
    reportNumbers := strings.Split(os.Getenv("REPORT_NUMBERS"), ";")
    logger.Debug("Getting report numbers", "report numbers", reportNumbers)
    c := cron.New()
    cronStr := "0 8 * * *" // Every day at 8:00 AM
    // cronStr := "*/5 * * * *"
    _, err = c.AddFunc(cronStr, func() { 
        err = reportService.SendDailyReport(reportNumbers)
        if err != nil {
            logger.Error("Cannot send daily report %w", err)
        }
    })
    if err != nil {
        log.Fatal("Error programando cron:", err)
    }
    c.Start()

	// Server
	apiPort := os.Getenv("API_PORT")
	host := fmt.Sprintf("%s:%s", "localhost", apiPort)
	server := pkg.NewServer(pkg.ServerOpts{
		ListenAddr: host,
		Logger:     logger,
	})

	go webhook.ConsumeEntries(commsService.NewCommunication)

	aircall := pkg.NewAircall(commsService.NewCommunication, logger)
	router.Handle("/aircall", aircall).Methods(http.MethodPost)

	router.HandleFunc("/pipedrive", handlers.HandleErrors(pipedriveApi.HandleOAuth)).Methods(http.MethodGet)

	router.HandleFunc("/webhooks", handlers.HandleErrors(webhook.ReciveNotificaction)).Methods(http.MethodPost)
	router.HandleFunc("/webhooks", handlers.HandleErrors(webhook.Verify)).Methods(http.MethodGet)

    // OneDrive images
    fs := http.StripPrefix("/onedrive/", http.FileServer(http.Dir("./onedrive/")))
    router.PathPrefix("/onedrive").Handler(fs)

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
