package whatsapp

import (
	"encoding/json"
	"fmt"
	"leadsextractor/models"
	"log/slog"
	"net/http"
)

type Webhook struct {
     entries    chan Entry 
     logger     *slog.Logger
}

// El payload que whatsapp nos envia
type NotificationPayload struct {
    Object      string   `json:"object"`
    Entries     []Entry  `json:"entry"`
}

type Entry struct {
    ID      string   `json:"id"`
    Changes []Change `json:"changes"`
}

type Change struct {
    Value Value  `json:"value"`
    Field string `json:"field"`
}

type Value struct {
    //MessagingProduct string   `json:"messaging_product"`
    Metadata         Metadata `json:"metadata"`
    Contacts         []Contact `json:"contacts"`
    Errors           []Error   `json:"errors"`
    //Messages          []Message `json:"messages"`
    Statuses         []Status  `json:"statuses"`
}

type Metadata struct {
    DisplayPhoneNumber string `json:"display_phone_number"`
    PhoneNumberID      string `json:"phone_number_id"`
}

type Error struct {
	// Common fields for both v15 and v16+
	Code  uint32    `json:"code"`
	Title string    `json:"title"`

	// Additional fields for v16+
	Message  string             `json:"message,omitempty"`
	ErrorData map[string]string `json:"error_data,omitempty"`
}

type Status struct {
    Id          uint32  `json:"id"`
    CustomerId  uint32  `json:"customer_id"` 
    Errors      []Error `json:"errors"`
    Status      string  `json:"status"` //delivered, read, sent
    Timestamp   uint64  `json:"timestamp"`
}

type Contact struct {
	WaID    string `json:"wa_id"`
	Profile struct {
        Name    string `json:"name"`
    }`json:"profile"`
}

func (wh *Webhook) ReciveNotificaction(w http.ResponseWriter, r *http.Request) error {
    defer r.Body.Close()
    var payload NotificationPayload
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        return err
    }

    for i, _ := range payload.Entries {
        wh.entries <- payload.Entries[i]
    }

    return nil
}

func (wh *Webhook) Verify(w http.ResponseWriter, r *http.Request) error {
    challenge := r.URL.Query().Get("hub.challenge")
    verify_token := r.URL.Query().Get("hub.verify_token")
    mode := r.URL.Query().Get("hub.mode")

    if mode == "subscribe" && verify_token == "Lautaro123." {
        w.Write([]byte(challenge))
    } else {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("400 - Bad request"))
    }
    
    return nil
}

func NewWebhook(l *slog.Logger) *Webhook{
    return &Webhook {
        entries: make(chan Entry),
        logger: l.With("module", "webhook"),
    }
}

func (wh *Webhook) ConsumeEntries(callback func (*models.Communication) error ) {
    for {
        select{
        case entry := <- wh.entries:
            wh.logger.Info("new notification recived")
            c, err := wh.Entry2Communication(&entry)
            if err != nil {
                continue
            }

            if err := callback(c); err != nil {
                wh.logger.Error(err.Error())
            }
        }
    }
}

func (wh *Webhook) Entry2Communication(e *Entry) (*models.Communication, error) {
    value := e.Changes[0].Value

    if len(value.Statuses) > 0 {
        for _, s := range value.Statuses {
            wh.logger.Debug("cambio de estado de un mensaje", "status", s.Status, "customerID", s.CustomerId)
        }
        return nil, nil
    }

    if len(value.Errors) > 0 {
        for _, e := range value.Errors {
            wh.logger.Error(e.Message, "code", e.Code, "title", e.Title)
        }
        return nil, fmt.Errorf("%s - %s", value.Errors[0].Code, value.Errors[0].Message)
    }

    wh.logger.Info("nuevo mensaje de whatsapp recibido", "phone", value.Contacts[0].WaID, "name", value.Contacts[0].Profile.Name)

    c := models.Communication {
        Fuente: "whatsapp",
        FechaLead: "",
        Fecha: "",
        Nombre: value.Contacts[0].Profile.Name,
        Link: fmt.Sprintf("https://web.whatsapp.com/send/?phone=%s", value.Contacts[0].WaID),
        Telefono: value.Contacts[0].WaID,
        Email: models.NullString{String: ""},
        Cotizacion: "",
        Asesor: models.Asesor{},
        Propiedad: models.Propiedad{},
        Busquedas: models.Busquedas{},
        IsNew: false,
    }

    return &c, nil
}
