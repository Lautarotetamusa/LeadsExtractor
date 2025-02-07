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
	"leadsextractor/infobip"
	"leadsextractor/logs"
	"leadsextractor/models"
	"leadsextractor/pipedrive"
	"leadsextractor/pkg"
	"leadsextractor/store"
	"leadsextractor/whatsapp"

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
	storer := store.NewStore(db, logger)

    leadStore := store.NewLeadStore(db)
    utmStore := store.NewUTMStore(db)
    commStore := store.NewCommStore(db)
    asesorStore := store.NewAsesorDBStore(db)

    // Round Robin
    var asesores []models.Asesor
    if err := asesorStore.GetAllActive(&asesores); err != nil{  
        log.Fatal("No se pudo obtener la lista de asesores")
    }
    rr := store.NewRoundRobin(asesores)

	flowManager := flow.NewFlowManager("actions.json", storer, logger)
	flow.DefineActions(wpp, pipedriveApi, infobipApi, leadStore)
	flowManager.MustLoad()

    // Services
    leadService := handlers.NewLeadService(leadStore)
    commsService := handlers.CommunicationService{ 
        RoundRobin: rr,
        Logger: logger,
        Flows: *flowManager,
        Store: *storer,

        Utms: utmStore,
        Comms: commStore,
        Leads: leadService,
    }
    asesorService := handlers.NewAsesorService(asesorStore, leadStore, rr)

    // Handlers
    leadHandler := handlers.NewLeadHandler(leadService)
    utmHandler := handlers.NewUTMHandler(utmStore)
	flowHandler := pkg.NewFlowHandler(flowManager)
    commHandler := handlers.NewCommHandler(commsService)
    asesorHandler := handlers.NewAsesorHandler(asesorService)

	router := mux.NewRouter()

    // Register routes
    leadHandler.RegisterRoutes(router)
    utmHandler.RegisterRoutes(router)
    flowHandler.RegisterRoutes(router)
    commHandler.RegisterRoutes(router)
    asesorHandler.RegisterRoutes(router)

    // Server
	apiPort := os.Getenv("API_PORT")
	host := fmt.Sprintf("%s:%s", "localhost", apiPort)
    server := pkg.NewServer(pkg.ServerOpts{
        ListenAddr: host,
        Logger: logger,
        FlowHandler: flowHandler,
        CommService: &commsService,
    })
	server.SetRoutes(router)

	go webhook.ConsumeEntries(commsService.NewCommunication)
	router.Use(CORS)

	router.HandleFunc("/pipedrive", pkg.HandleErrors(pipedriveApi.HandleOAuth)).Methods("GET")

	router.HandleFunc("/wame", pkg.HandleErrors(whatsapp.GenerateWppLink)).Methods("POST", "OPTIONS")
	router.HandleFunc("/encode", pkg.HandleErrors(whatsapp.GenerateEncodeMsg)).Methods("POST", "OPTIONS")
	router.HandleFunc("/webhooks", pkg.HandleErrors(webhook.ReciveNotificaction)).Methods("POST", "OPTIONS")
	router.HandleFunc("/webhooks", pkg.HandleErrors(webhook.Verify)).Methods("GET", "OPTIONS")

	// Logs
	router.HandleFunc("/logs", pkg.HandleErrors(logsHandler.GetLogs)).Methods("GET", "OPTIONS")

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
