package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"leadsextractor/messenger/flow"
	"leadsextractor/messenger/store"
	"log/slog"
	"text/template"
	"time"

	"github.com/google/uuid"
)

type IncomingPayload struct {
	Phone   string `json:"phone"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Source  string `json:"source"`
	IsNew   bool   `json:"is_new"`
}

type Service struct {
	store  store.Storer
	flows  *flow.Manager
	logger *slog.Logger
}

func New(s store.Storer, fm *flow.Manager, l *slog.Logger) *Service {
	return &Service{store: s, flows: fm, logger: l.With("module", "messenger")}
}

func (s *Service) HandleIncoming(p *IncomingPayload) error {
	now := time.Now()
	incoming := &store.Message{
		Phone:       p.Phone,
		Type:        "incoming",
		Content:     p.Content,
		Outgoing:    false,
		ScheduledAt: now,
		SendedAt:    &now,
	}
	if err := s.store.Insert(incoming); err != nil {
		return fmt.Errorf("error guardando mensaje entrante: %w", err)
	}

	flowUUID := s.flows.MainUUID
	if last, err := s.store.GetLastOutgoingWithResponse(p.Phone); err == nil && last.OnResponse.Valid {
		flowUUID = last.OnResponse.UUID
	}

	return s.runFlow(p, flowUUID)
}

func (s *Service) ScheduleMessage(phone, msgType, content string, scheduledAt time.Time, onResponse *uuid.UUID) error {
	msg := &store.Message{
		Phone:       phone,
		Type:        msgType,
		Content:     content,
		Outgoing:    true,
		ScheduledAt: scheduledAt,
	}
	if onResponse != nil {
		msg.OnResponse = uuid.NullUUID{UUID: *onResponse, Valid: true}
	}
	return s.store.Insert(msg)
}

func (s *Service) runFlow(p *IncomingPayload, flowUUID uuid.UUID) error {
	f, err := s.flows.Get(flowUUID)
	if err != nil {
		return err
	}

	for _, rule := range f.Rules {
		if !rule.Condition.Matches(p.Source, p.Content, p.IsNew) {
			continue
		}
		for _, action := range rule.Actions {
			if err := s.scheduleAction(p, action, rule.OnResponse); err != nil {
				s.logger.Error("error programando accion", "action", action.Name, "err", err)
			}
		}
		break
	}
	return nil
}

func (s *Service) scheduleAction(p *IncomingPayload, action flow.Action, onResponse uuid.NullUUID) error {
	content, msgType, err := renderAction(action, p)
	if err != nil {
		return err
	}

	scheduledAt := time.Now().Add(time.Duration(action.Interval))
	msg := &store.Message{
		Phone:       p.Phone,
		Type:        msgType,
		Content:     content,
		Outgoing:    true,
		ScheduledAt: scheduledAt,
		OnResponse:  onResponse,
	}
	return s.store.Insert(msg)
}

type wppMessageParams struct {
	Text string `json:"text"`
}

func renderAction(action flow.Action, p *IncomingPayload) (content, msgType string, err error) {
	switch action.Name {
	case "wpp.message":
		var params wppMessageParams
		if err = json.Unmarshal(action.Params, &params); err != nil {
			return
		}
		content, err = renderTemplate(params.Text, p)
		msgType = "wpp.message"
	case "wpp.template":
		content = string(action.Params)
		msgType = "wpp.template"
	default:
		err = fmt.Errorf("accion no soportada en messenger: %s", action.Name)
	}
	return
}

func renderTemplate(tmpl string, p *IncomingPayload) (string, error) {
	t, err := template.New("msg").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, p); err != nil {
		return "", err
	}
	return buf.String(), nil
}
