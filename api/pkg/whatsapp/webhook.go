package whatsapp

import (
	"encoding/json"
	"fmt"
	"leadsextractor/models"
	"log/slog"
	"net/http"
	"net/url"
)

type Webhook struct {
	entries chan Entry
	token   string
	logger  *slog.Logger
}

// El payload que whatsapp nos envia
type NotificationPayload struct {
	Object  string  `json:"object"`
	Entries []Entry `json:"entry"`
}

type NewLinkPayload struct {
	To     string            `json:"to"`
	Msg    string            `json:"msg"`
	Params map[string]string `json:"params"`
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
	Metadata Metadata  `json:"metadata"`
	Contacts []Contact `json:"contacts"`
	Errors   []Error   `json:"errors"`
	Messages []Message `json:"messages"`
	Statuses []Status  `json:"statuses"`
}

type Message struct {
	From      string      `json:"from"`
	Id        string      `json:"id"`
	Timestamp string      `json:"timestamp"`
	Type      MessageType `json:"type"`

	Text        *Text        `json:"text,omitempty"`
	Interactive *Interactive `json:"interactive,omitempty"`
	Button      *Button      `json:"button,omitempty"`
}

type Text struct {
	Body string `json:"body"`
}

type Button struct {
	Text    string `json:"text"`
	Payload string `json:"payload"`
}

type Interactive struct {
	ButtonReply *ButtonReply `json:"button_reply,omitempty"`
	ListReply   *ListReply   `json:"list_reply,omitempty"`
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
	Code  uint32 `json:"code"`
	Title string `json:"title"`

	// Additional fields for v16+
	Message   string            `json:"message,omitempty"`
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
		Name string `json:"name"`
	} `json:"profile"`
}

func NewWebhook(token string, l *slog.Logger) *Webhook {
	return &Webhook{
		entries: make(chan Entry),
		logger:  l.With("module", "webhook"),
		token:   token,
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
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		wh.logger.Error("error parseando el payload", "err", err.Error())
		return err
	}

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

func (wh *Webhook) ConsumeEntries(callback func(*models.Communication) error) {
	for {
		select {
		case entry := <-wh.entries:
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
