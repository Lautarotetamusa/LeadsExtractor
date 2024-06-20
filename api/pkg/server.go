package pkg

import (
	"fmt"
	"leadsextractor/models"
	"leadsextractor/store"
	"log"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type Server struct {
    listenAddr  string
    roundRobin  *store.RoundRobin
    Store       *store.Store
    logger      *slog.Logger

    flowHandler *FlowHandler
}
type HandlerErrorFn func(w http.ResponseWriter, r *http.Request) error
type HandlerFn func(w http.ResponseWriter, r *http.Request)

func NewServer(listenAddr string, logger *slog.Logger, db *sqlx.DB, fh *FlowHandler) *Server {
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
		Store:      s,
        logger:     logger,
        flowHandler: fh,
    }
}

func (s *Server) SetRoutes(router *mux.Router) {
	router.Use(CORS)

	router.HandleFunc("/asesor", HandleErrors(s.GetAllAsesores)).Methods("GET")
	router.HandleFunc("/asesor/{phone}", HandleErrors(s.GetOneAsesor)).Methods("GET")
	router.HandleFunc("/asesor", HandleErrors(s.InsertAsesor)).Methods("POST")
	router.HandleFunc("/asesor/{phone}", HandleErrors(s.UpdateAsesor)).Methods("PUT")
	router.HandleFunc("/asesores/", HandleErrors(s.UpdateStatuses)).Methods("POST", "OPTIONS")

	router.HandleFunc("/assign", HandleErrors(s.AssignAsesor)).Methods("POST")


	router.HandleFunc("/lead", HandleErrors(s.GetAll)).Methods("GET")
	router.HandleFunc("/lead/{phone}", HandleErrors(s.GetOne)).Methods("GET")
	router.HandleFunc("/lead", HandleErrors(s.Insert)).Methods("POST")
	router.HandleFunc("/lead/{phone}", HandleErrors(s.Update)).Methods("PUT")

	router.HandleFunc("/communication", HandleErrors(s.HandleNewCommunication)).Methods("POST")
	router.HandleFunc("/communications", HandleErrors(s.GetCommunications)).Methods("GET", "OPTIONS")

	router.HandleFunc("/broadcast", HandleErrors(s.NewBroadcast)).Methods("POST")
	router.HandleFunc("/mainFlow", HandleErrors(s.flowHandler.SetFlowAsMain)).Methods("POST")
}

func (s *Server) Run(router *mux.Router) {
	s.logger.Info(fmt.Sprintf("Server started at %s", s.listenAddr))
	if err := http.ListenAndServe(s.listenAddr, router); err != nil {
		s.logger.Error("No se pudo iniciar el server", err)
	}
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
