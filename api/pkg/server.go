package pkg

import (
	"fmt"
	"leadsextractor/infobip"
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
}
type HandlerErrorFn func (w http.ResponseWriter, r *http.Request) (error)
type HandlerFn func (w http.ResponseWriter, r *http.Request) 

func NewServer(listenAddr string, db *sqlx.DB, roundRobin *RoundRobin, infobipApi *infobip.InfobipApi) *Server{
    return &Server{
        listenAddr: listenAddr,
        db: db,
        roundRobin: roundRobin,
        infobipApi: infobipApi,
    }
}

func (s *Server) Run(){
    asesorHandler := AsesorHandler{
        Store: &store.AsesorMysqlStorage{
            Db: s.db,
        },
    } 

    leadHandler := LeadHandler{
        Store: &store.LeadMySqlStorage{
            Db: s.db,
        },
    } 

    router := mux.NewRouter()

    router.Use(enableCORS)

    router.HandleFunc("/asesor", handleErrors(asesorHandler.GetAll)).Methods("GET")
    router.HandleFunc("/asesor/{phone}", handleErrors(asesorHandler.GetOne)).Methods("GET")
    router.HandleFunc("/asesor", handleErrors(asesorHandler.Insert)).Methods("POST")
    router.HandleFunc("/asesor/{phone}", handleErrors(asesorHandler.Update)).Methods("PUT")

    router.HandleFunc("/lead", handleErrors(leadHandler.GetAll)).Methods("GET")
    router.HandleFunc("/lead/{phone}", handleErrors(leadHandler.GetOne)).Methods("GET")
    router.HandleFunc("/lead", handleErrors(leadHandler.Insert)).Methods("POST")
    router.HandleFunc("/lead/{phone}", handleErrors(leadHandler.Update)).Methods("PUT")

    //router.HandleFunc("/communication", s.HandleCors).Methods("OPTIONS")
    router.HandleFunc("/communication", handleErrors(s.HandleNewCommunication)).Methods("POST")
    router.HandleFunc("/communication", handleErrors(s.HandleListCommunication)).Methods("GET")
    router.HandleFunc("/communication", s.HandleCors).Methods("OPTIONS")

    server := http.Server{
        Addr: s.listenAddr,
        Handler: router,
    };

    fmt.Println("Server started at %s", server.Addr)
    err := server.ListenAndServe()
    if err != nil {
        log.Fatal("No se pudo iniciar el server\n", err)
    }
    fmt.Println("Server started at %s", server.Addr)
}

func enableCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        //w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
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
