package pkg

import (
	"encoding/json"
	"errors"
	"leadsextractor/models"
	"leadsextractor/numbers"
	"net/http"

	"github.com/dinistavares/go-aircall-api"
)

func (s *Server) AircallWebhook(w http.ResponseWriter, r *http.Request) {
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
        s.logger.Error("error parsing the payload", "err", err.Error(), "module", "aircall")
        return 
    }

    s.logger.Info("webhook payload received", "event", payload.Event, "module", "aircall")
    // Only save it when the call is created
    if payload.Event != "call.created" {
        return
    }

    call, err := payload.GetCallData()
    if err != nil {
        s.logger.Error(err.Error())
        return 
    }

    communication, err := callToCommunication(call)
    if err != nil {
        s.logger.Error(err.Error(), "module", "aircall")
        return 
    }

    err = s.NewCommunication(communication)
    if err != nil {
        s.logger.Error(err.Error(), "module", "aircall")
        return 
    }
}

func callToCommunication(call *aircall.Call) (*models.Communication, error) {
    contact := call.Contact

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
