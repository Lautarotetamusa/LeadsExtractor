package service

import (
	"context"
	"encoding/json"
	"leadsextractor/messenger/store"
	"leadsextractor/pkg/whatsapp"
	"log/slog"
	"time"
)

type Scheduler struct {
	store    store.Storer
	wpp      *whatsapp.Whatsapp
	interval time.Duration
	logger   *slog.Logger
}

func NewScheduler(s store.Storer, wpp *whatsapp.Whatsapp, interval time.Duration, l *slog.Logger) *Scheduler {
	return &Scheduler{
		store:    s,
		wpp:      wpp,
		interval: interval,
		logger:   l.With("module", "scheduler"),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	s.logger.Info("scheduler iniciado", "interval", s.interval)
	for {
		select {
		case <-ticker.C:
			s.process()
		case <-ctx.Done():
			s.logger.Info("scheduler detenido")
			return
		}
	}
}

func (s *Scheduler) process() {
	msgs, err := s.store.GetPending()
	if err != nil {
		s.logger.Error("error obteniendo mensajes pendientes", "err", err)
		return
	}

	for _, msg := range msgs {
		if err := s.send(msg); err != nil {
			s.logger.Error("error enviando mensaje", "id", msg.Id, "phone", msg.Phone, "err", err)
			continue
		}
		if err := s.store.MarkSent(msg.Id); err != nil {
			s.logger.Error("error marcando mensaje como enviado", "id", msg.Id, "err", err)
		}
	}
}

func (s *Scheduler) send(msg *store.Message) error {
	switch msg.Type {
	case "wpp.message":
		_, err := s.wpp.SendMessage(msg.Phone, msg.Content)
		return err
	case "wpp.template":
		var t whatsapp.TemplatePayload
		if err := json.Unmarshal([]byte(msg.Content), &t); err != nil {
			return err
		}
		_, err := s.wpp.SendTemplate(msg.Phone, t)
		return err
	default:
		s.logger.Warn("tipo de mensaje desconocido, ignorando", "type", msg.Type, "id", msg.Id)
		return nil
	}
}
