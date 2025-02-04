package pkg

import (
	"leadsextractor/store"
	"net/http"
	"reflect"
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

// 2 de enero del 2006, se usa siempre esta fecha por algun motivo extraño xd
const timeLayout string = "2006-01-02"
var timeConverter = func(value string) reflect.Value {
    if v, err := time.Parse(timeLayout, value); err == nil {
        return reflect.ValueOf(v)
    }
    return reflect.Value{} // this is the same as the private const invalidType
}
