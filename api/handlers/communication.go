package handlers

import (
	"encoding/json"
	"fmt"
	"leadsextractor/middleware"
	"leadsextractor/models"
	"leadsextractor/store"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

type CommunicationHandler struct {
	service CommunicationService
}

const maxUploadCsvComms = 200

func NewCommHandler(s CommunicationService) *CommunicationHandler {
	return &CommunicationHandler{
		service: s,
	}
}

func (h CommunicationHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/communication").Subrouter()
    r.Methods(http.MethodOptions)

	r.HandleFunc("", HandleErrors(h.Insert)).Methods(http.MethodPost)
	r.HandleFunc("", HandleErrors(h.GetAll)).Methods(http.MethodGet)

    var comms []models.Communication
    csvHandler := middleware.NewCSVHandler("csv_file", "comms", comms).
        WithLimit(maxUploadCsvComms)

	r.Handle("/csv", csvHandler.CSVMiddleware(
        http.HandlerFunc(HandleErrors(h.HandleCSVUpload)),
    )).Methods(http.MethodPost)
}

func (h CommunicationHandler) GetAll(w http.ResponseWriter, r *http.Request) error {
	var (
		params  store.QueryParam
		decoder = schema.NewDecoder()
		count   int
	)

	decoder.RegisterConverter(time.Time{}, timeConverter)
	if err := decoder.Decode(&params, r.URL.Query()); err != nil {
		return err
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
			Page:     params.Page,
			PageSize: params.PageSize,
			Items:    len(communications),
			Total:    count,
		},
		Data: communications,
	}
	json.NewEncoder(w).Encode(res)
	return nil
}

func (h CommunicationHandler) Insert(w http.ResponseWriter, r *http.Request) error {
	var c models.Communication
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		return jsonErr(err)
	}

	if err := h.service.NewCommunication(&c); err != nil {
		return err
	}

	createdResponse(w, "communication created succesfuly", c)
	return nil
}

func (h CommunicationHandler) HandleCSVUpload(w http.ResponseWriter, r *http.Request) error {
	var (
		errorCount, successCount int
		errors                   []MultipleError
		wg                       sync.WaitGroup
	)
    comms, ok := r.Context().Value("comms").([]models.Communication)
    if !ok {
        return ErrBadRequest("comms does not exists in the context")
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
		"total_error_count":   errorCount,
		"errors":              errors,
	})

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
