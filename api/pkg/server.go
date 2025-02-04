package pkg

import (
	"fmt"
	"leadsextractor/handlers"
	"leadsextractor/store"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
    listenAddr  string
    roundRobin  *store.RoundRobin
    logger      *slog.Logger

    flowHandler *FlowHandler
    Store       *store.Store
    leads   *store.LeadStore
    comms   *handlers.CommunicationService
}

type ServerOpts struct {
    ListenAddr  string
    RoundRobin  *store.RoundRobin
    Logger      *slog.Logger

    FlowHandler *FlowHandler
    Store       *store.Store
    LeadStore   *store.LeadStore
    CommService *handlers.CommunicationService
}

func NewServer(opts ServerOpts) *Server {
	return &Server{
		listenAddr: opts.ListenAddr,
		roundRobin: opts.RoundRobin,
		Store:      opts.Store,
        comms:      opts.CommService,
        leads:      opts.LeadStore,
        logger:     opts.Logger,
        flowHandler: opts.FlowHandler,
    }
}

func (s *Server) SetRoutes(router *mux.Router) {
	router.HandleFunc("/asesor", HandleErrors(s.GetAllAsesores)).Methods("GET", "OPTIONS")
	router.HandleFunc("/asesor/{phone}", HandleErrors(s.GetOneAsesor)).Methods("GET", "OPTIONS")
	router.HandleFunc("/asesor", HandleErrors(s.InsertAsesor)).Methods("POST", "OPTIONS")
	router.HandleFunc("/asesor/{phone}", HandleErrors(s.UpdateAsesor)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/asesor/{phone}", HandleErrors(s.DeleteAsesor)).Methods("DELETE", "OPTIONS")
    router.HandleFunc("/asesor/{phone}/reasign", HandleErrors(s.Reasign)).Methods("PUT", "OPTIONS")

	router.HandleFunc("/broadcast", HandleErrors(s.NewBroadcast)).Methods("POST", "OPTIONS")
	router.HandleFunc("/mainFlow", HandleErrors(s.flowHandler.SetFlowAsMain)).Methods("POST", "OPTIONS")

	router.HandleFunc("/metrics", HandleErrors(s.HandleMetrics)).Methods("GET", "OPTIONS")

    aircall := NewAircall(s.comms.NewCommunication, s.logger)
    // aircall := AircallTestHandler{}
    router.Handle("/aircall", aircall).Methods("POST", "OPTIONS")
}

func (s *Server) Run(router *mux.Router) {
	s.logger.Info(fmt.Sprintf("Server started at %s", s.listenAddr))
	if err := http.ListenAndServe(s.listenAddr, router); err != nil {
		s.logger.Error("No se pudo iniciar el server", "err", err.Error())
	}
}
