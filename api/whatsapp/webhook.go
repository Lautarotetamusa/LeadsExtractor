package whatsapp

import (
	"encoding/json"
	"fmt"
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
    err := json.NewDecoder(r.Body).Decode(&payload)
    if err != nil {
        return err
    }

    for i, _ := range payload.Entries {
        wh.entries <- payload.Entries[i]
    }

    return nil
}

func (wh *Webhook) Verify(w http.ResponseWriter, r *http.Request) error {
    challenge := r.URL.Query().Get("hub.challenge")
    w.Write([]byte(challenge))
    return nil
}

func NewWebhook(l *slog.Logger) *Webhook{
    return &Webhook {
        entries: make(chan Entry),
        logger: l.With("module", "webhook"),
    }
}

func (wh *Webhook) ConsumeEntries(callback func (*Entry) error ) {
    select{
    case entry := <- wh.entries:
        wh.logger.Info("new notification recived")
        if err := callback(&entry); err != nil {
            wh.logger.Error(err.Error())
        }
    default:
    }
}

func HandleNotification(e *Entry) error {
    fmt.Printf("%v\n", e)
    return nil
}
