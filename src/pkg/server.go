package pkg

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	listenAddr string
	logger     *slog.Logger
}

type ServerOpts struct {
	ListenAddr string
	Logger     *slog.Logger
}

func NewServer(opts ServerOpts) *Server {
	return &Server{
		listenAddr: opts.ListenAddr,
		logger:     opts.Logger,
	}
}

func (s *Server) Run(router *mux.Router) {
	s.logger.Info(fmt.Sprintf("Server started at %s", s.listenAddr))
	if err := http.ListenAndServe(s.listenAddr, router); err != nil {
		s.logger.Error("No se pudo iniciar el server", "err", err.Error())
	}
}
