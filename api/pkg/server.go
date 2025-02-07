package pkg

import (
	"fmt"
	"leadsextractor/handlers"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
    listenAddr  string
    logger      *slog.Logger

    flowHandler *FlowHandler
    comms   *handlers.CommunicationService
}

type ServerOpts struct {
    ListenAddr  string
    Logger      *slog.Logger

    FlowHandler *FlowHandler
    CommService *handlers.CommunicationService
}

func NewServer(opts ServerOpts) *Server {
	return &Server{
		listenAddr: opts.ListenAddr,
        comms:      opts.CommService,
        logger:     opts.Logger,
        flowHandler: opts.FlowHandler,
    }
}

func (s *Server) SetRoutes(router *mux.Router) {
	router.HandleFunc("/broadcast", HandleErrors(s.NewBroadcast)).Methods("POST", "OPTIONS")

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
