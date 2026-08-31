package flow

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Interval time.Duration

type Condition struct {
	IsNew   *bool  `json:"is_new,omitempty"`
	Source  string `json:"source,omitempty"`
	Message string `json:"message,omitempty"`
}

type Action struct {
	Name     string          `json:"action"`
	Interval Interval        `json:"interval"`
	Params   json.RawMessage `json:"params"`
}

type Rule struct {
	Condition  Condition     `json:"condition"`
	Actions    []Action      `json:"actions"`
	OnResponse uuid.NullUUID `json:"on_response,omitempty"`
}

type Flow struct {
	Rules []Rule `json:"rules"`
	Name  string `json:"name"`
}

type Manager struct {
	MainUUID uuid.UUID          `json:"Main"`
	Flows    map[uuid.UUID]Flow `json:"Flows"`
}

func (c *Condition) Matches(source, message string, isNew bool) bool {
	if c.IsNew != nil && *c.IsNew != isNew {
		return false
	}
	if c.Source != "" && !strings.EqualFold(c.Source, source) {
		return false
	}
	if c.Message != "" && !strings.Contains(strings.ToUpper(message), strings.ToUpper(c.Message)) {
		return false
	}
	return true
}

func (i *Interval) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*i = Interval(d)
	return nil
}

func (i Interval) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(i).String())
}
