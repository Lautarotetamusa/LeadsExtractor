package store

import (
	"fmt"
	"leadsextractor/pkg/numbers"
	"strings"

	"github.com/google/uuid"
)

// Un flow tiene muchos actions
type ActionSave struct {
	Name       string        `db:"name"` // Alguno de los definidos en DefineActions
	Nro        int           `db:"nro"`  // Posicion en el flow
	FlowUUID   uuid.UUID     `db:"flow_uuid"`
	SendedAt   string        `db:"sended_at"`
	LeadPhone  string        `db:"lead_phone"`
	// Flow que vamos a ejecutar en caso de que alguien responda a este action  
	OnResponse uuid.NullUUID `db:"on_response,omitempty"`
}

var insertFieldsAction = []string{"name", "nro", "flow_uuid", "lead_phone", "on_response"}

const tableNameAction = "Action"

func (s *Store) SaveAction(a *ActionSave) error {
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableNameAction,
		strings.Join(insertFieldsAction, ", "),
		":"+strings.Join(insertFieldsAction, ", :"))

	if _, err := s.db.NamedExec(query, a); err != nil {
		return err
	}
	return nil
}

func (s *Store) GetLastActionFromLead(leadPhone numbers.PhoneNumber) (*ActionSave, error) {
	var a ActionSave
	query := fmt.Sprintf("SELECT * FROM %s WHERE lead_phone=? ORDER BY sended_at DESC LIMIT 1", tableNameAction)
	if err := s.db.Get(&a, query, leadPhone.String()); err != nil {
		return nil, err
	}
	return &a, nil
}
