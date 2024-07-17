package pkg

import (
	"encoding/json"
	"fmt"
	"leadsextractor/models"
	"leadsextractor/store"
	"net/http"
	"reflect"
	"time"

	"github.com/gorilla/schema"
)

type ListResponse = struct {
    Success     bool                    `json:"success"`
    Pagination  Pagination              `json:"pagination"`
    Data        []models.Communication  `json:"data"`
}

type Pagination struct{
    Page        int `json:"page"`
    PageSize    int `json:"page_size"`
    Items       int `json:"items"`
    Total       int `json:"total"`
}

// 2 de enero del 2006, se usa siempre esta fecha por algun motivo extraño xd
const timeLayout string = "2006-01-02"
var timeConverter = func(value string) reflect.Value {
    if v, err := time.Parse(timeLayout, value); err == nil {
        return reflect.ValueOf(v)
    }
    return reflect.Value{} // this is the same as the private const invalidType
}

var boolConverter = func(value string) reflect.Value {
    fmt.Println(value)
    switch value{
    case "true":
        fmt.Println("true")
        return reflect.ValueOf(true)
    case "false":
        fmt.Println("is false")
        return reflect.ValueOf(false)
    default:
        return reflect.Value{}
    }
}

func (s *Server) GetCommunications(w http.ResponseWriter, r *http.Request) error {
    var (
        params store.QueryParam
        decoder = schema.NewDecoder()
        count int
    )

    decoder.RegisterConverter(time.Time{}, timeConverter)
    decoder.RegisterConverter(true, boolConverter)

    if err := decoder.Decode(&params, r.URL.Query()); err != nil{
        return err;
    }
    fmt.Printf("%#v", params)

    communications, err := s.Store.GetCommunications(&params) 
    if err != nil {
        return err
    }

    count, err = s.Store.GetCommunicationsCount(&params)
    if err != nil {
        return err
    }

    w.Header().Set("Content-Type", "application/json")
    res := ListResponse{
        Success: true,
        Pagination: Pagination{
            Page: params.Page,
            PageSize: params.PageSize,
            Items: len(communications),
            Total: count,
        },
        Data: communications,
    }
    json.NewEncoder(w).Encode(res)
	return nil
}

func (s *Server) NewCommunication(c *models.Communication) error {
	source, err := s.Store.GetSource(c)
	if err != nil {
		return err
	}

	lead, err := s.Store.InsertOrGetLead(s.roundRobin, c)
	if err != nil {
		return err
	}

	c.Asesor = lead.Asesor
    c.Cotizacion = lead.Cotizacion
    c.Email = lead.Email
    c.Nombre = lead.Name

    go s.flowHandler.manager.RunMainFlow(c)
        
    if err = s.Store.InsertCommunication(c, source); err != nil {
        s.logger.Error(err.Error(), "path", "InsertCommunication")
        return err
    }
    return nil
}

func (s *Server) HandleNewCommunication(w http.ResponseWriter, r *http.Request) error {
	c := &models.Communication{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(c); err != nil {
		return err
	}
    
    if err := s.NewCommunication(c); err != nil {
        return err
    }

    successResponse(w, "communication created succesfuly", c)
	return nil
}
