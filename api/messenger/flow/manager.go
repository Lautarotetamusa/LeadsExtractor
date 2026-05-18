package flow

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
)

func Load(path string) (*Manager, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir el archivo de flows: %w", err)
	}
	defer f.Close()

	var m Manager
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return nil, fmt.Errorf("no se pudo parsear el archivo de flows: %w", err)
	}
	return &m, nil
}

func (m *Manager) Get(id uuid.UUID) (*Flow, error) {
	flow, ok := m.Flows[id]
	if !ok {
		return nil, fmt.Errorf("flow %s no encontrado", id)
	}
	return &flow, nil
}

func (m *Manager) GetMain() (*Flow, error) {
	return m.Get(m.MainUUID)
}
