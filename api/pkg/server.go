package pkg

import (
	"fmt"
	"leadsextractor/infobip"
	"leadsextractor/pipedrive"
	"leadsextractor/store"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	listenAddr string
	roundRobin *store.RoundRobin
	infobipApi *infobip.InfobipApi
	pipedrive  *pipedrive.Pipedrive
	Store      *store.Store
}
type HandlerErrorFn func(w http.ResponseWriter, r *http.Request) error
type HandlerFn func(w http.ResponseWriter, r *http.Request)

func NewServer(listenAddr string, db *sqlx.DB, infobipApi *infobip.InfobipApi, pipedrive *pipedrive.Pipedrive) *Server {
	s := store.NewStore(db)
	return &Server{
		listenAddr: listenAddr,
		roundRobin: store.NewRoundRobin(s),
		infobipApi: infobipApi,
		pipedrive:  pipedrive,
		Store:      s,
	}
}

func (s *Server) Run() {
	router := mux.NewRouter()

	router.Use(CORS)

	router.HandleFunc("/asesor", handleErrors(s.GetAllAsesores)).Methods("GET")
	router.HandleFunc("/asesor/{phone}", handleErrors(s.GetOneAsesor)).Methods("GET")
	router.HandleFunc("/asesor", handleErrors(s.InsertAsesor)).Methods("POST")
	router.HandleFunc("/asesor/{phone}", handleErrors(s.UpdateAsesor)).Methods("PUT")

	router.HandleFunc("/asesores/", handleErrors(s.UpdateStatuses)).Methods("POST", "OPTIONS")

	router.HandleFunc("/lead", handleErrors(s.GetAll)).Methods("GET")
	router.HandleFunc("/lead/{phone}", handleErrors(s.GetOne)).Methods("GET")
	router.HandleFunc("/lead", handleErrors(s.Insert)).Methods("POST")
	router.HandleFunc("/lead/{phone}", handleErrors(s.Update)).Methods("PUT")

	router.HandleFunc("/communication", handleErrors(s.HandleNewCommunication)).Methods("POST")
	router.HandleFunc("/assign", handleErrors(s.AssignAsesor)).Methods("POST")
	router.HandleFunc("/communication", handleErrors(s.HandleListCommunication)).Methods("GET", "OPTIONS")

	router.HandleFunc("/pipedrive", handleErrors(s.HandlePipedriveOAuth)).Methods("GET")

	fmt.Printf("Server started at %s\n", s.listenAddr)
	err := http.ListenAndServe(s.listenAddr, router)
	if err != nil {
		log.Fatal("No se pudo iniciar el server\n", err)
	}
	fmt.Printf("Server started at %s\n", s.listenAddr)
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
