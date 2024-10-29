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
	"leadsextractor/infobip"
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

    w := os.Stderr

    logger := slog.New(
        tint.NewHandler(w, &tint.Options{
            Level:      slog.LevelDebug,
            TimeFormat: time.DateTime,
        }),
    )

	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	db := store.ConnectDB(ctx)

	apiPort := os.Getenv("API_PORT")
	host := fmt.Sprintf("%s:%s", "localhost", apiPort)

	infobipApi := infobip.NewInfobipApi(
        os.Getenv("INFOBIP_APIURL"),
        os.Getenv("INFOBIP_APIKEY"),
        "5213328092850",
        logger,
    )

    pipedriveConfig := pipedrive.Config{
        ClientId: os.Getenv("PIPEDRIVE_CLIENT_ID"),
        ClientSecret: os.Getenv("PIPEDRIVE_CLIENT_SECRET"),
        ApiToken: os.Getenv("PIPEDRIVE_API_TOKEN"),
        RedirectURI: os.Getenv("PIPEDRIVE_REDIRECT_URI"),
    }
	pipedriveApi := pipedrive.NewPipedrive(pipedriveConfig, logger)

	wpp := whatsapp.NewWhatsapp(
        os.Getenv("WHATSAPP_ACCESS_TOKEN"),

        os.Getenv("WHATSAPP_NUMBER_ID"),
        logger,
	)

    storer := store.NewStore(db, logger)

    flowManager := flow.NewFlowManager("actions.json", storer, logger)
    flow.DefineActions(wpp, pipedriveApi, infobipApi, storer)

    flowManager.MustLoad()

    webhook := whatsapp.NewWebhook(
        os.Getenv("WHATSAPP_VERIFY_TOKEN"),
        logger,
    )

	router := mux.NewRouter()

    router.Use(loggingMiddleware)

    flowHandler := pkg.NewFlowHandler(flowManager)

	server := pkg.NewServer(host, logger, db, flowHandler)
    server.SetRoutes(router)
    
    go webhook.ConsumeEntries(server.NewCommunication)
	router.Use(CORS)

	router.HandleFunc("/pipedrive", pkg.HandleErrors(pipedriveApi.HandleOAuth)).Methods("GET")

    router.HandleFunc("/wame", pkg.HandleErrors(whatsapp.GenerateWppLink)).Methods("POST", "OPTIONS")
    router.HandleFunc("/encode", pkg.HandleErrors(whatsapp.GenerateEncodeMsg)).Methods("POST", "OPTIONS")
    router.HandleFunc("/webhooks", pkg.HandleErrors(webhook.ReciveNotificaction)).Methods("POST", "OPTIONS")
    router.HandleFunc("/webhooks", pkg.HandleErrors(webhook.Verify)).Methods("GET", "OPTIONS")

	router.HandleFunc("/actions", pkg.HandleErrors(flowHandler.GetConfig)).Methods("GET", "OPTIONS")

	router.HandleFunc("/flows", pkg.HandleErrors(flowHandler.NewFlow)).Methods("POST", "OPTIONS")
	router.HandleFunc("/flows/{uuid}", pkg.HandleErrors(flowHandler.UpdateFlow)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/flows", pkg.HandleErrors(flowHandler.GetFlows)).Methods("GET", "OPTIONS")
	router.HandleFunc("/flows/main", pkg.HandleErrors(flowHandler.GetMainFlow)).Methods("GET", "OPTIONS")
	router.HandleFunc("/flows/{uuid}", pkg.HandleErrors(flowHandler.GetFlow)).Methods("GET", "OPTIONS")
	router.HandleFunc("/flows/{uuid}", pkg.HandleErrors(flowHandler.DeleteFlow)).Methods("DELETE", "OPTIONS")

	server.Run(router)
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        //log.Println(r.Method, r.RequestURI)
        next.ServeHTTP(w, r)
    })
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
