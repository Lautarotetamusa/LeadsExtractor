package whatsapp

import (
	"encoding/json"
	"fmt"
	"leadsextractor/models"
	"leadsextractor/numbers"
	"log/slog"
	"net/http"
	"net/url"
)

type Webhook struct {
     entries    chan Entry 
     token      string
     logger     *slog.Logger
}

// El payload que whatsapp nos envia
type NotificationPayload struct {
    Object      string   `json:"object"`
    Entries     []Entry  `json:"entry"`
}

type NewLinkPayload struct {
    To      string            `json:"to"`
    Msg     string            `json:"msg"`
    Params  map[string]string `json:"params"` 
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
    Messages         []Message `json:"messages"`
    Statuses         []Status  `json:"statuses"`
}

type Message struct {
    From        string  `json:"from"`
    Id          string  `json:"id"`
    Timestamp   string  `json:"timestamp"`
    Type        string  `json:"type"`
    Text        struct {
        Body    string  `json:"body"`
    } `json:"text,omitempty"`
    Interactive Interactive `json:"interactive,omitempty"`
}

type Interactive struct {
    Type InteractiveType `json:"type"`
}

type InteractiveType struct {
    ButtonReply ButtonReply `json:"button_reply,omitempty"`
    ListReply   ListReply   `json:"list_reply,omitempty"`
}

type ButtonReply struct {
    ID    string `json:"id"`
    Title string `json:"title"`
}

type ListReply struct {
    ID          string `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
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
    Id          string  `json:"id"`
    RecipientId string  `json:"recipient_id"` 
    Errors      []Error `json:"errors"`
    Status      string  `json:"status"` //delivered, read, sent
    Timestamp   string  `json:"timestamp"`
}

type Contact struct {
	WaID    string `json:"wa_id"`
	Profile struct {
        Name    string `json:"name"`
    }`json:"profile"`
}

func NewWebhook(token string, l *slog.Logger) *Webhook{
    return &Webhook {
        entries: make(chan Entry),
        logger: l.With("module", "webhook"),
        token: token,
    }
}

func GenerateEncodeMsg(w http.ResponseWriter, r *http.Request) error {
    defer r.Body.Close()
    var payload NewLinkPayload
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        return err
    }

    params := url.Values{}
    for k, v := range payload.Params {
        params.Add(k, v)
    }

    encodedParams := encodeString(params.Encode())
    w.Write([]byte(encodedParams))
    return nil
}

func GenerateWppLink(w http.ResponseWriter, r *http.Request) error {
    defer r.Body.Close()
    var payload NewLinkPayload
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        return err
    }

    params := url.Values{}
    for k, v := range payload.Params {
        params.Add(k, v)
    }

    encodedParams := url.QueryEscape(encodeString(params.Encode()))
    link := fmt.Sprintf(baseLinkUrl, payload.To, encodedParams, payload.Msg)
    w.Write([]byte(link))
    return nil
}

func (wh *Webhook) ReciveNotificaction(w http.ResponseWriter, r *http.Request) error {
    defer r.Body.Close()
    var payload NotificationPayload
    // var data interface{}
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        wh.logger.Error("error parseando el payload", "err", err.Error())
        return err
    }

    // j, _ := json.MarshalIndent(data, " ", "\t")
    // fmt.Println(string(j))

    for _, e := range payload.Entries {
        wh.entries <- e
    }

    w.WriteHeader(http.StatusOK)
    return nil
}

func (wh *Webhook) Verify(w http.ResponseWriter, r *http.Request) error {
    challenge := r.URL.Query().Get("hub.challenge")
    verify_token := r.URL.Query().Get("hub.verify_token")
    mode := r.URL.Query().Get("hub.mode")

    if mode == "subscribe" && verify_token == wh.token {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(challenge))
    } else {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("400 - Bad request"))
    }
    
    return nil
}

func (wh *Webhook) ConsumeEntries(callback func (*models.Communication) error ) {
    for {
        select{
        case entry := <- wh.entries:
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
    if (len(e.Changes) <= 0) {
        return nil, fmt.Errorf("whatsapp entry doesn't have any Changes")
    }
    value := e.Changes[0].Value

    if len(value.Statuses) > 0 {
        for _, s := range value.Statuses {
            wh.logger.Debug("cambio de estado de un mensaje", 
                "status", s.Status, 
                "recipientId", s.RecipientId, 
                "msgId", s.Id)
            for _, e := range s.Errors {
                wh.logger.Error(e.Message, "code", e.Code, "to", s.RecipientId)
            }
        }
        return nil, fmt.Errorf("cambio de estado")
    }

    if len(value.Errors) > 0 {
        for _, e := range value.Errors {
            wh.logger.Error(e.Message, "code", e.Code, "title", e.Title)
        }
        return nil, fmt.Errorf("%d - %s", value.Errors[0].Code, value.Errors[0].Message)
    }

    if (len(value.Contacts) <= 0) {
        return nil, fmt.Errorf("whatsapp value doesn't have any Contacts")
    }
    if (len(value.Messages) <= 0) {
        return nil, fmt.Errorf("whatsapp value doesn't have any Messages")
    }

    // TODO: Leer todos los mensajes? por qué hay muchos contactos?
    contact := value.Contacts[0]
    message := value.Messages[0]

    wh.logger.Info("nuevo mensaje de whatsapp recibido", 
        "phone", contact.WaID, 
        "name", contact.Profile.Name, 
        "id", message.Id)
    
    phone, err := numbers.NewPhoneNumber(contact.WaID)
    if err != nil {
        return nil, fmt.Errorf("error parsing whatsapp number: %s", contact.WaID)
    }

    c := models.Communication {
        Fuente: "whatsapp",
        FechaLead: "",
        Fecha: "",
        Nombre: contact.Profile.Name,
        Link: fmt.Sprintf("https://web.whatsapp.com/send/?phone=%s", contact.WaID),
        Telefono: *phone,
        Email: models.NullString{String: ""},
        Cotizacion: "",
        Asesor: models.Asesor{},
        Propiedad: models.Propiedad{},
        Busquedas: models.Busquedas{},
        IsNew: false, // WHAT?
        Message: models.NullString{String: message.Text.Body, Valid: true},
        Wamid: models.NullString{String: message.Id, Valid: true},
    }

    return &c, nil
}
