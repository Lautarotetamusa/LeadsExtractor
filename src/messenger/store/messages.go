package store

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Message struct {
	Id          int64         `db:"id"`
	Phone       string        `db:"phone"`
	Type        string        `db:"type"`
	Content     string        `db:"content"`
	Outgoing    bool          `db:"outgoing"`
	ScheduledAt time.Time     `db:"scheduled_at"`
	SendedAt    *time.Time    `db:"sended_at"`
	OnResponse  uuid.NullUUID `db:"on_response"`
}

type Storer interface {
	Insert(m *Message) error
	GetPending() ([]*Message, error)
	MarkSent(id int64) error
	GetLastOutgoingWithResponse(phone string) (*Message, error)
}

type Store struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Insert(m *Message) error {
	query := `INSERT INTO Messages (phone, type, content, outgoing, scheduled_at, on_response)
	          VALUES (:phone, :type, :content, :outgoing, :scheduled_at, :on_response)`
	_, err := s.db.NamedExec(query, m)
	return err
}

func (s *Store) GetPending() ([]*Message, error) {
	var msgs []*Message
	query := `SELECT * FROM Messages
	          WHERE outgoing = TRUE AND scheduled_at <= NOW() AND sended_at IS NULL
	          ORDER BY scheduled_at ASC
	          LIMIT 100`
	if err := s.db.Select(&msgs, query); err != nil {
		return nil, fmt.Errorf("error obteniendo mensajes pendientes: %w", err)
	}
	return msgs, nil
}

func (s *Store) MarkSent(id int64) error {
	_, err := s.db.Exec(`UPDATE Messages SET sended_at = NOW() WHERE id = ?`, id)
	return err
}

func (s *Store) GetLastOutgoingWithResponse(phone string) (*Message, error) {
	var msg Message
	query := `SELECT * FROM Messages
	          WHERE phone = ? AND outgoing = TRUE AND on_response IS NOT NULL AND sended_at IS NOT NULL
	          ORDER BY sended_at DESC LIMIT 1`
	if err := s.db.Get(&msg, query, phone); err != nil {
		return nil, err
	}
	return &msg, nil
}
