package handler

import (
	"encoding/json"
	"leadsextractor/messenger/service"
	"leadsextractor/pkg/whatsapp"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handler struct {
	svc     *service.Service
	webhook *whatsapp.Webhook
	logger  *slog.Logger
}

func New(svc *service.Service, webhook *whatsapp.Webhook, l *slog.Logger) *Handler {
	return &Handler{svc: svc, webhook: webhook, logger: l.With("module", "handler")}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/incoming", h.handleIncoming).Methods(http.MethodPost)
	r.HandleFunc("/webhooks", h.handleWhatsappWebhook).Methods(http.MethodPost)
	r.HandleFunc("/webhooks", h.handleWhatsappVerify).Methods(http.MethodGet)
	r.HandleFunc("/messages/schedule", h.handleSchedule).Methods(http.MethodPost)
	r.HandleFunc("/messages/pending", h.handlePending).Methods(http.MethodGet)
}

// POST /incoming — payload de portales o scrapers
func (h *Handler) handleIncoming(w http.ResponseWriter, r *http.Request) {
	var p service.IncomingPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.svc.HandleIncoming(&p); err != nil {
		h.logger.Error("error procesando incoming", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// POST /webhooks — WhatsApp webhook
func (h *Handler) handleWhatsappWebhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var payload whatsapp.NotificationPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, entry := range payload.Entries {
		p, err := entryToIncoming(&entry)
		if err != nil {
			h.logger.Debug("entrada de whatsapp ignorada", "err", err)
			continue
		}
		if err := h.svc.HandleIncoming(p); err != nil {
			h.logger.Error("error procesando mensaje de whatsapp", "err", err)
		}
	}
	w.WriteHeader(http.StatusOK)
}

// GET /webhooks — verificación de WhatsApp
func (h *Handler) handleWhatsappVerify(w http.ResponseWriter, r *http.Request) {
	h.webhook.Verify(w, r)
}

type scheduleRequest struct {
	Phone       string     `json:"phone"`
	Type        string     `json:"type"`
	Content     string     `json:"content"`
	ScheduledAt time.Time  `json:"scheduled_at"`
	OnResponse  *uuid.UUID `json:"on_response,omitempty"`
}

// POST /messages/schedule — programar un mensaje futuro
func (h *Handler) handleSchedule(w http.ResponseWriter, r *http.Request) {
	var req scheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Phone == "" || req.Content == "" {
		http.Error(w, "phone y content son requeridos", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		req.Type = "wpp.message"
	}
	if err := h.svc.ScheduleMessage(req.Phone, req.Type, req.Content, req.ScheduledAt, req.OnResponse); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// GET /messages/pending — ver mensajes en cola
func (h *Handler) handlePending(w http.ResponseWriter, r *http.Request) {
	// delegado al store directamente a través del service en una próxima iteración
	w.WriteHeader(http.StatusNotImplemented)
}

func entryToIncoming(e *whatsapp.Entry) (*service.IncomingPayload, error) {
	if len(e.Changes) == 0 {
		return nil, errNoChanges
	}
	value := e.Changes[0].Value
	if len(value.Messages) == 0 || len(value.Contacts) == 0 {
		return nil, errNoMessage
	}
	msg := value.Messages[0]
	contact := value.Contacts[0]

	content := ""
	if msg.Text != nil {
		content = msg.Text.Body
	} else if msg.Button != nil {
		content = msg.Button.Text
	} else if msg.Interactive != nil && msg.Interactive.ButtonReply != nil {
		content = msg.Interactive.ButtonReply.Title
	}

	return &service.IncomingPayload{
		Phone:   contact.WaID,
		Name:    contact.Profile.Name,
		Content: content,
		Source:  "whatsapp",
	}, nil
}

var (
	errNoChanges = httpError("entrada sin changes")
	errNoMessage = httpError("entrada sin mensajes")
)

type httpError string

func (e httpError) Error() string { return string(e) }
