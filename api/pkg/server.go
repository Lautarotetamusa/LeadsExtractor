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
    leadStore   *store.LeadStore
    logger      *slog.Logger

    flowHandler *FlowHandler
}
type HandlerErrorFn func(w http.ResponseWriter, r *http.Request) error
type HandlerFn func(w http.ResponseWriter, r *http.Request)

func NewServer(
    listenAddr string, 
    logger *slog.Logger, 
    db *sqlx.DB, 
    fh *FlowHandler,
    leadStore   *store.LeadStore,
) *Server {
	s := store.NewStore(db, logger)

    var asesores []models.Asesor
    if err := s.GetAllActiveAsesores(&asesores); err != nil{  
        log.Fatal("No se pudo obtener la lista de asesores")
    }
    rr := store.NewRoundRobin(asesores)

	return &Server{
		listenAddr: listenAddr,
		roundRobin: rr,
		Store:      s,
        leadStore: leadStore,
        logger:     logger,
        flowHandler: fh,
    }
}

func (s *Server) SetRoutes(router *mux.Router) {
	router.HandleFunc("/asesor", HandleErrors(s.GetAllAsesores)).Methods("GET", "OPTIONS")
	router.HandleFunc("/asesor/{phone}", HandleErrors(s.GetOneAsesor)).Methods("GET", "OPTIONS")
	router.HandleFunc("/asesor", HandleErrors(s.InsertAsesor)).Methods("POST", "OPTIONS")
	router.HandleFunc("/asesor/{phone}", HandleErrors(s.UpdateAsesor)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/asesor/{phone}", HandleErrors(s.DeleteAsesor)).Methods("DELETE", "OPTIONS")
    router.HandleFunc("/asesor/{phone}/reasign", HandleErrors(s.Reasign)).Methods("PUT", "OPTIONS")

	router.HandleFunc("/utm", HandleErrors(s.GetAllUtmes)).Methods("GET", "OPTIONS")
	router.HandleFunc("/utm/{id}", HandleErrors(s.GetOneUtm)).Methods("GET", "OPTIONS")
	router.HandleFunc("/utm", HandleErrors(s.InsertUtm)).Methods("POST", "OPTIONS")
	router.HandleFunc("/utm/{id}", HandleErrors(s.UpdateUtm)).Methods("PUT", "OPTIONS")

	router.HandleFunc("/communication", HandleErrors(s.HandleNewCommunication)).Methods("POST", "OPTIONS")
	router.HandleFunc("/communications", HandleErrors(s.GetCommunications)).Methods("GET", "OPTIONS")
	router.HandleFunc("/communication-csv", HandleErrors(s.HandleCSVUpload)).Methods("POST", "OPTIONS")

	router.HandleFunc("/broadcast", HandleErrors(s.NewBroadcast)).Methods("POST", "OPTIONS")
	router.HandleFunc("/mainFlow", HandleErrors(s.flowHandler.SetFlowAsMain)).Methods("POST", "OPTIONS")

	router.HandleFunc("/metrics", HandleErrors(s.HandleMetrics)).Methods("GET", "OPTIONS")

    aircall := NewAircall(s.NewCommunication, s.logger)
    // aircall := AircallTestHandler{}
    router.Handle("/aircall", aircall).Methods("POST", "OPTIONS")
}

func (s *Server) Run(router *mux.Router) {
	s.logger.Info(fmt.Sprintf("Server started at %s", s.listenAddr))
	if err := http.ListenAndServe(s.listenAddr, router); err != nil {
		s.logger.Error("No se pudo iniciar el server", "err", err.Error())
	}
}
