package pkg

import (
	"fmt"
	"leadsextractor/infobip"
	"leadsextractor/models"
	"leadsextractor/pipedrive"
	"leadsextractor/store"
	"leadsextractor/whatsapp"
	"log"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type Server struct {
    listenAddr  string
    roundRobin  *store.RoundRobin
    infobipApi  *infobip.InfobipApi
    pipedrive   *pipedrive.Pipedrive
    whatsapp    *whatsapp.Whatsapp
    Store       *store.Store
    logger      *slog.Logger
}
type HandlerErrorFn func(w http.ResponseWriter, r *http.Request) error
type HandlerFn func(w http.ResponseWriter, r *http.Request)

func NewServer(
    listenAddr string, 
    logger *slog.Logger,
    db *sqlx.DB, 
    infobipApi *infobip.InfobipApi, 
    pipedrive *pipedrive.Pipedrive,
    whatsapp *whatsapp.Whatsapp,
) *Server {
	s := store.NewStore(db, logger)
    var asesores []models.Asesor
    err := s.GetAllActiveAsesores(&asesores)

    if err != nil{  
        log.Fatal("No se pudo obtener la lista de asesores")
    }
    rr := store.NewRoundRobin(asesores)

	return &Server{
		listenAddr: listenAddr,
		roundRobin: rr,
		infobipApi: infobipApi,
		pipedrive:  pipedrive,
		Store:      s,
        logger:     logger,
        whatsapp:   whatsapp,
    }
}

func (s *Server) Run() {
	router := mux.NewRouter()

	router.Use(CORS)

    s.setupActions()

	router.HandleFunc("/asesor", handleErrors(s.GetAllAsesores)).Methods("GET")
	router.HandleFunc("/asesor/{phone}", handleErrors(s.GetOneAsesor)).Methods("GET")
	router.HandleFunc("/asesor", handleErrors(s.InsertAsesor)).Methods("POST")
	router.HandleFunc("/asesor/{phone}", handleErrors(s.UpdateAsesor)).Methods("PUT")
	router.HandleFunc("/assign", handleErrors(s.AssignAsesor)).Methods("POST")

	router.HandleFunc("/asesores/", handleErrors(s.UpdateStatuses)).Methods("POST", "OPTIONS")

	router.HandleFunc("/lead", handleErrors(s.GetAll)).Methods("GET")
	router.HandleFunc("/lead/{phone}", handleErrors(s.GetOne)).Methods("GET")
	router.HandleFunc("/lead", handleErrors(s.Insert)).Methods("POST")
	router.HandleFunc("/lead/{phone}", handleErrors(s.Update)).Methods("PUT")

	router.HandleFunc("/communication", handleErrors(s.NewCommunication)).Methods("POST")
	router.HandleFunc("/communications", handleErrors(s.GetCommunications)).Methods("GET", "OPTIONS")


	router.HandleFunc("/actions", handleErrors(NewAction)).Methods("POST", "OPTIONS")
	router.HandleFunc("/actions", handleErrors(s.GetFlows)).Methods("GET")

	router.HandleFunc("/broadcast", handleErrors(s.NewBroadcast)).Methods("POST")
	router.HandleFunc("/mainFlow", handleErrors(s.SetFlowAsMain)).Methods("POST")

	router.HandleFunc("/pipedrive", handleErrors(s.pipedrive.HandleOAuth)).Methods("GET")

	s.logger.Info(fmt.Sprintf("Server started at %s", s.listenAddr))
	err := http.ListenAndServe(s.listenAddr, router)
	if err != nil {
		s.logger.Error("No se pudo iniciar el server\n", err)
	}
	s.logger.Info(fmt.Sprintf("Server started at %s\n", s.listenAddr))
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
func handleErrors(fn HandlerErrorFn) HandlerFn {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			ErrorResponse(w, r, err)
		}
	}
}
