package pkg

import (
	"leadsextractor/store"
	"net/http"
	"time"

	"github.com/gorilla/schema"
)

func (s *Server) HandleMetrics(w http.ResponseWriter, r *http.Request) error {
    var (
        params store.QueryParam
        decoder = schema.NewDecoder()
    )

    decoder.RegisterConverter(time.Time{}, timeConverter)
    if err := decoder.Decode(&params, r.URL.Query()); err != nil{
        return err;
    }

	metric, err := s.Store.GetMetrics(&params)
	if err != nil {
		return err
	}

	dataResponse(w, metric)
	return nil
}
