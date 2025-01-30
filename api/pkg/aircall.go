package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"leadsextractor/models"
	"leadsextractor/numbers"
	"net/http"
	"log/slog"

	"github.com/dinistavares/go-aircall-api"
)

type AircallWebhookHandler struct {
    saveFunc SaveCommFunc
    logger *slog.Logger
}

type SaveCommFunc func(c *models.Communication) error

func NewAircall(saveFunc SaveCommFunc, logger *slog.Logger) *AircallWebhookHandler {
    return &AircallWebhookHandler {
        saveFunc: saveFunc,
        logger: logger.With("module", "aircall"),
    }
}

type AircallTestHandler struct {}

func (h AircallWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    // https://developer.aircall.io/api-references/#webhook-usage
    // The server must always return a 2XX code to prevent aircall from blocking the webhook
    w.WriteHeader(http.StatusOK)

    defer r.Body.Close()
    var payload aircall.InboundWebhook
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        h.logger.Error("error parsing the payload", "err", err.Error())
        return 
    }

    h.logger.Info("webhook payload received", "event", payload.Event)
    // Only save it when the call is created
    if payload.Event != "call.created" {
        return
    }

    call, err := payload.GetCallData()
    if err != nil {
        h.logger.Error(err.Error())
        return 
    }

    communication, err := callToCommunication(call)
    if err != nil {
        h.logger.Error(err.Error())
        return 
    }

    err = h.saveFunc(communication)
    if err != nil {
        h.logger.Error(err.Error())
        return 
    }
}

func (h AircallTestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    w.WriteHeader(http.StatusOK)

    defer r.Body.Close()
    var res map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
        return 
    }

    output, err := json.MarshalIndent(res, " ", "\t")
    if err != nil {
        return
    }

    fmt.Printf(string(output))
}

func callToCommunication(call *aircall.Call) (*models.Communication, error) {
    if call == nil {
        return nil, errors.New("call is nil")
    }
    contact := call.Contact
    if contact == nil {
        return nil, errors.New("contact is nil")
    }

    if len(*contact.PhoneNumbers) <= 0 {
        return nil, errors.New("call contact doesnt have phone number")
    }

    phone, err := numbers.NewPhoneNumber((*contact.PhoneNumbers)[0].Value)
    if err != nil {
        return nil, err
    }

    // If the contact doenst have name, set the phone as the name
    name := contact.FirstName + " " + contact.LastName
    if name == " " {
        name = phone.String()
    }

    comm := &models.Communication{
        Fuente: "ivr",
        Nombre: name,
        Telefono: *phone,
    }

    if len(*contact.Emails) > 0 {
        comm.Email = models.NullString{String: (*contact.Emails)[0].Value, Valid: true}
    }

    return comm, nil
}
