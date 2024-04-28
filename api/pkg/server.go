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

type Server struct{
    listenAddr  string
    db          *sqlx.DB
    roundRobin  *RoundRobin
    infobipApi  *infobip.InfobipApi
    pipedrive   *pipedrive.Pipedrive
    asesorHandler *AsesorHandler
}
type HandlerErrorFn func (w http.ResponseWriter, r *http.Request) (error)
type HandlerFn func (w http.ResponseWriter, r *http.Request) 

func NewServer(listenAddr string, db *sqlx.DB, roundRobin *RoundRobin, infobipApi *infobip.InfobipApi, pipedrive *pipedrive.Pipedrive) *Server{
    return &Server{
        listenAddr: listenAddr,
        db: db,
        roundRobin: roundRobin,
        infobipApi: infobipApi,
        pipedrive:  pipedrive,
        asesorHandler: &AsesorHandler{
            Store: &store.AsesorMysqlStorage{
                Db: db,
            },
        },
    }
}

func (s *Server) Run(){
    leadHandler := LeadHandler{
        Store: &store.LeadMySqlStorage{
            Db: s.db,
        },
    } 
    router := mux.NewRouter()

    router.Use(CORS)

    router.HandleFunc("/asesor", handleErrors(s.asesorHandler.GetAll)).Methods("GET")
    router.HandleFunc("/asesor/{phone}", handleErrors(s.asesorHandler.GetOne)).Methods("GET")
    router.HandleFunc("/asesor", handleErrors(s.asesorHandler.Insert)).Methods("POST")
    router.HandleFunc("/asesor/{phone}", handleErrors(s.asesorHandler.Update)).Methods("PUT")

    router.HandleFunc("/asesores/", handleErrors(s.UpdateStatuses)).Methods("POST", "OPTIONS")

    router.HandleFunc("/lead", handleErrors(leadHandler.GetAll)).Methods("GET")
    router.HandleFunc("/lead/{phone}", handleErrors(leadHandler.GetOne)).Methods("GET")
    router.HandleFunc("/lead", handleErrors(leadHandler.Insert)).Methods("POST")
    router.HandleFunc("/lead/{phone}", handleErrors(leadHandler.Update)).Methods("PUT")

    router.HandleFunc("/communication", handleErrors(s.HandleNewCommunication)).Methods("POST")
    router.HandleFunc("/communication", handleErrors(s.HandleListCommunication)).Methods("GET", "OPTIONS")

    router.HandleFunc("/pipedrive", handleErrors(s.HandlePipedriveOAuth)).Methods("GET")

    fmt.Println("Server started at %s",s.listenAddr)
    err := http.ListenAndServe(s.listenAddr, router)
    if err != nil {
        log.Fatal("No se pudo iniciar el server\n", err)
    }
    fmt.Println("Server started at %s", s.listenAddr)
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
func handleErrors(fn HandlerErrorFn) (HandlerFn) {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := fn(w, r); err != nil{
            ErrorResponse(w, r, err)
        }
    }
}
