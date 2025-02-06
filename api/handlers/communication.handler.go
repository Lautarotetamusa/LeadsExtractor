package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"leadsextractor/flow"
	"leadsextractor/models"
	"leadsextractor/store"
	"log/slog"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

type CommunicationHandler struct {
    service CommunicationService
}

type CommunicationService struct {
    RoundRobin  *store.RoundRobin
    Logger  *slog.Logger

    Leads   LeadService

    Flows   flow.FlowManager
    Utms    store.UTMStorer
    Comms   store.CommunicationStorer
    Store   store.Store
}

const maxUploadCsvComms = 200

var requiredCSVHeaders = []string{"fuente", "fecha", "propiedad.titulo", "propiedad.url", "propiedad.id", "propiedad.precio", "propiedad.ubicacion", "propiedad.habitaciones", "propiedad.banios", "propiedad.area", "nombre", "telefono", "email", "mensaje"}

func NewCommHandler(s CommunicationService) *CommunicationHandler {
    return &CommunicationHandler{
        service: s,
    }
}

func (h CommunicationHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/communication", HandleErrors(h.Insert)).Methods("POST", "OPTIONS")
	router.HandleFunc("/communications", HandleErrors(h.GetAll)).Methods("GET", "OPTIONS")
	router.HandleFunc("/communication-csv", HandleErrors(h.HandleCSVUpload)).Methods("POST", "OPTIONS")
}

func (s CommunicationService) StoreCommunication(c *models.Communication) error {
	source, err := s.Store.GetSource(c)
	if err != nil {
		return err
	}

	lead, err := s.Leads.GetOrInsert(s.RoundRobin, c)
	if err != nil {
		return err
	}

    s.findUtmInMessage(c)
    
	c.Asesor = lead.Asesor
    c.Cotizacion = lead.Cotizacion
    c.Email = lead.Email
    c.Nombre = lead.Name

    if err = s.Comms.Insert(c, source); err != nil {
        s.Logger.Error(err.Error(), "path", "InsertCommunication")
        return err
    }
    if c.Message.Valid && c.Message.String != "" {
        if err = s.Store.InsertMessage(store.CommunicationToMessage(c)); err != nil {
            return err
        }
    }
    return nil
}

func (s CommunicationService) NewCommunication(c *models.Communication) error {
    err := s.StoreCommunication(c)
    if err != nil {
        return err
    }
    
    s.runAction(c)

    return nil
}

func (h CommunicationHandler) GetAll(w http.ResponseWriter, r *http.Request) error {
    var (
        params store.QueryParam
        decoder = schema.NewDecoder()
        count int
    )

    decoder.RegisterConverter(time.Time{}, timeConverter)
    if err := decoder.Decode(&params, r.URL.Query()); err != nil{
        return err;
    }

    communications, err := h.service.Comms.GetAll(&params) 
    if err != nil {
        return err
    }

    count, err = h.service.Comms.Count(&params)
    if err != nil {
        return err
    }

    w.Header().Set("Content-Type", "application/json")
    res := ListResponse{
        Success: true,
        Pagination: store.Pagination{
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

func (h CommunicationHandler) Insert(w http.ResponseWriter, r *http.Request) error {
    var c models.Communication
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(c); err != nil {
		return err
	}
    
    if err := h.service.NewCommunication(&c); err != nil {
        return err
    }

    successResponse(w, "communication created succesfuly", c)
	return nil
}

func getCommunicationsFromCSV(r *http.Request, comms *[]models.Communication) error {
    file, _, err := r.FormFile("csv_file")
    if err != nil {
        return fmt.Errorf("error reading the csv_file %v", err.Error())
    }
    defer file.Close()

    // Read the headers with encoding/csv
    csvReader := csv.NewReader(file)
    headers, err := csvReader.Read()
    if err != nil {
        return fmt.Errorf("error reading the CSV headers: %v", err)
    }

    if err := csvValidateHeaders(headers); err != nil {
        return err
    }

    // Come back to the start of the file for the Unmarshall to work
    file.Seek(0, io.SeekStart)

    if err := gocsv.UnmarshalMultipartFile(&file, comms); err != nil {
        return err
    }

    if len(*comms) > maxUploadCsvComms {
        return fmt.Errorf("no se pueden cargar mas de %d comunicaciones de una sola vez", maxUploadCsvComms)
    }
    return nil
}

func csvValidateHeaders(headers []string) error {
    missingFields := make([]string, 0)
    for _, header := range requiredCSVHeaders {
        if !slices.Contains(headers, header) {
            missingFields = append(missingFields, header)
        }
    }
    if len(missingFields) > 0 {
        return fmt.Errorf("fields %s are missing", strings.Join(missingFields, ", "))
    }
    return nil
}

func (h CommunicationHandler) HandleCSVUpload(w http.ResponseWriter, r *http.Request) error {
    var (
        comms []models.Communication
        errorCount, successCount int
        errors []MultipleError
        wg sync.WaitGroup
    )
    if err := getCommunicationsFromCSV(r, &comms); err != nil {
        return err
    }

    // double check to dont create a channel with a massive length
    if len(comms) > maxUploadCsvComms {
        panic("too many communications in the csv file")
    }

    errorSet := make(map[string]int)
    errChan := make(chan error, len(comms))

    // Add utm_source with the current date
    utm_source := fmt.Sprintf("csv_file_%s", time.Now().Format(time.DateOnly))

    // Launch goroutines for each communication
    for _, c := range comms {
        wg.Add(1)
        go func(comm models.Communication) {
            defer wg.Done()
            comm.Utm.Source = models.NullString{String: utm_source, Valid: true}

            if err := h.service.StoreCommunication(&comm); err != nil {
                errChan <- err
            }         
        }(c)
    }

    // CLose the results after all goroutines finish
    go func() {
		wg.Wait()
		close(errChan)
	}()

    for err := range errChan {
        errorSet[err.Error()] += 1
        errorCount += 1
    }

    successCount = len(comms) - errorCount

    status := http.StatusOK
    if errorCount > 0 {
        errors = collectMultipleErrors(errorSet)
        status = http.StatusMultipleChoices
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]any{
        "total_success_count": successCount,
        "total_error_count": errorCount,
        "errors": errors,
    })

    return nil
}

func (s CommunicationService) runAction(c *models.Communication) {
    action, err := s.Store.GetLastActionFromLead(c.Telefono)
    if err != nil {
        s.Logger.Error("cannot get the lead last action", "err", err)
    }

    if action != nil && action.OnResponse.Valid {
        go s.Flows.RunFlow(c, action.OnResponse.UUID)
    }else{
        go s.Flows.RunMainFlow(c)
    }
}

// find if the message have match with any utm
func (s CommunicationService) findUtmInMessage(c *models.Communication) {
    var Utms []models.UtmDefinition
    err := s.Utms.GetAll(&Utms)
    if err != nil {
        s.Logger.Error(err.Error())
        return 
    }
    if !c.Message.Valid {
        return 
    }

    message := strings.ToUpper(c.Message.String)
    for _, utm := range Utms {
        if strings.Contains(message, utm.Code) {
            s.Logger.Info(fmt.Sprintf("found code %s in message", utm.Code))
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

// 2 de enero del 2006, se usa siempre esta fecha por algun motivo extraño xd
const timeLayout string = "2006-01-02"
var timeConverter = func(value string) reflect.Value {
    if v, err := time.Parse(timeLayout, value); err == nil {
        return reflect.ValueOf(v)
    }
    return reflect.Value{} // this is the same as the private const invalidType
}
