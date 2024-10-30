package pkg

import (
	"encoding/json"
	"fmt"
	"leadsextractor/models"
	"leadsextractor/store"
	"net/http"
	"reflect"
	"strings"
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

func (s *Server) GetCommunications(w http.ResponseWriter, r *http.Request) error {
    var (
        params store.QueryParam
        decoder = schema.NewDecoder()
        count int
    )

    decoder.RegisterConverter(time.Time{}, timeConverter)
    if err := decoder.Decode(&params, r.URL.Query()); err != nil{
        return err;
    }

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

    s.findUtmInMessage(c)
    
	c.Asesor = lead.Asesor
    c.Cotizacion = lead.Cotizacion
    c.Email = lead.Email
    c.Nombre = lead.Name

    action, err := s.Store.GetLastActionFromLead(c.Telefono)
    // flow, _ := s.flowHandler.manager.GetFlow(a.FlowUUID)
    // action := flow.Rules[0].Actions[a.Nro]
    if err != nil {
        s.logger.Error("cannot get the lead last action", "err", err)
    }

    if action != nil && action.OnResponse.Valid {
        go s.flowHandler.manager.RunFlow(c, action.OnResponse.UUID)
    }else{
        go s.flowHandler.manager.RunMainFlow(c)
    }

        
    if err = s.Store.InsertCommunication(c, source); err != nil {
        s.logger.Error(err.Error(), "path", "InsertCommunication")
        return err
    }
    if len(c.Message) > 0 {
        if err = s.Store.InsertMessage(store.CommunicationToMessage(c)); err != nil {
            return err
        }
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

// TODO: los utms se podrían guardar en Server
func (s *Server) findUtmInMessage(c *models.Communication) {
    var utms []models.UtmDefinition
    err := s.Store.GetAllUtm(&utms)
    if err != nil {
        s.logger.Error(err.Error())
        return 
    }
    if c.Message == "" {
        return 
    }

    message := strings.ToUpper(c.Message)
    for _, utm := range utms {
        if strings.Contains(message, utm.Code) {
            s.logger.Info(fmt.Sprintf("found code %s in message", utm.Code))
            c.Utm = models.Utm{
                Medium: utm.Medium,
                Source: utm.Source,
                Campaign: utm.Campaign,
                Ad: utm.Ad,
                Channel: utm.Channel,
            }
        }
    }
}
