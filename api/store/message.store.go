package store

import (
	"fmt"
	"leadsextractor/models"
	"strings"
)

type Message struct {
    Id      int `db:"id"`
    IdCommunication     int `db:"id_communication"`
    CreatedAt   string `db:"created_at"`
    Text    string  `db:"text"`
    Wamid   string  `db:"wamid"`
}
var insertFields = []string{"id_communication", "text", "wamid"};
const tableName = "Message";

func CommunicationToMessage(c *models.Communication) *Message {
    return &Message{
        IdCommunication: c.Id,
        Text: c.Message.String,
    }
}

func (s *Store) InsertMessage(m *Message) error {
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
        tableName,
        strings.Join(insertFields, ", "), 
        ":"+strings.Join(insertFields, ", :"))

	if _, err := s.db.NamedExec(query, m); err != nil {
		return err
	}
	return nil
}
