package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"time"

	"leadsextractor/flow"
	"leadsextractor/handlers"
	"leadsextractor/pkg"
	"leadsextractor/pkg/infobip"
	"leadsextractor/pkg/logs"
	"leadsextractor/pkg/pipedrive"
	"leadsextractor/pkg/roundrobin"
	"leadsextractor/pkg/whatsapp"
	"leadsextractor/store"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/lmittmann/tint"
)

func main() {
	ctx := context.Background()

	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	//TODO: Agregar un index a campo 'time' de la db de logs
	client := logs.MongoConnect(ctx, &logs.MongoConnectionSettings{
		User: os.Getenv("DB_USER"),
		Pass: os.Getenv("DB_PASS"),
		Host: os.Getenv("HOST"),
		Port: int16(27017),
	})
	collection := client.Database("db").Collection("log")

	w := os.Stdout
	mw := logs.NewMongoWriter(collection)

	logger := slog.New(
		logs.Fanout(
			slog.NewJSONHandler(mw, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}),
			tint.NewHandler(w, &tint.Options{
				Level:      slog.LevelDebug,
				TimeFormat: time.DateTime,
			}),
		),
	)

	logsHandler := logs.NewLogsHandler(ctx, collection)

	db := store.ConnectDB(ctx)

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
    propHandler := handlers.NewPropertyHandler(propStore)
    publishPropHandler := handlers.NewPublishedPropertyHandler(publisPropStore)
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

	// Logs
	router.HandleFunc("/logs", handlers.HandleErrors(logsHandler.GetLogs)).Methods(http.MethodGet, http.MethodOptions)

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
