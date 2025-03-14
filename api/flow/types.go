package flow

import (
	"encoding/json"
	"fmt"
	"leadsextractor/models"
	"leadsextractor/pkg/whatsapp"
	"leadsextractor/store"
	"log/slog"
	"reflect"
	"time"

	"github.com/google/uuid"
)

type Interval time.Duration

type ActionFunc func(c *models.Communication, params interface{}) error

type ActionDefinition struct {
	Func      ActionFunc
	ParamType reflect.Type
}

type SendWppTextParam struct {
	Text string `json:"text"`
}

type SendWppMedia struct {
	Image *whatsapp.MediaPayload `json:"image,omitempty"`
	Video *whatsapp.MediaPayload `json:"video,omitempty"`
}

type Action struct {
	Name     string      `json:"action"`
	Interval Interval    `json:"interval"`
	Params   interface{} `json:"params"`
}

type Rule struct {
	Condition  store.QueryParam `json:"condition"`
	Actions    []Action         `json:"actions"`
	OnResponse uuid.NullUUID    `json:"on_response,omitempty"` // Si es nulo significa que debemos ejecutar el flow main
}

type Flow struct {
	Rules []Rule
	Name  string
}

type FlowManager struct {
	Main     uuid.UUID //Es el uuid del flow que se ejecuta cuando llega una comuncacion
	Flows    map[uuid.UUID]Flow
	filename string
	logger   *slog.Logger
	storer   *store.Store
}

// Son las acciones permitidas
var actions map[string]ActionDefinition

func DefineAction(name string, f ActionFunc, t reflect.Type) {
	actions[name] = ActionDefinition{
		Func:      f,
		ParamType: t,
	}
}

func NewFlowManager(filename string, s *store.Store, l *slog.Logger) *FlowManager {
	actions = make(map[string]ActionDefinition)
	return &FlowManager{
		filename: filename,
		logger:   l.With("module", "flow"),
		Flows:    make(map[uuid.UUID]Flow),
		storer:   s,
	}
}

func (a *Action) UnmarshalJSON(data []byte) error {
	var temp struct {
		Name     string          `json:"action"`
		Interval Interval        `json:"interval"`
		Params   json.RawMessage `json:"params"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	a.Name = temp.Name
	a.Interval = temp.Interval

	actionDef, exists := actions[a.Name]
	if !exists {
		return fmt.Errorf("la accion %s no existe", a.Name)
	}

	//Deserealizamos al tipo de esta accion definido en ParamsType
	if actionDef.ParamType == nil {
		a.Params = nil
		return nil
	}
	params := reflect.New(actionDef.ParamType).Interface()
	if err := json.Unmarshal(temp.Params, params); err != nil {
		return fmt.Errorf("parámetros inválidos para la acción %s: %v", a.Name, err)
	}
	a.Params = params

	return nil
}

func (i *Interval) UnmarshalJSON(data []byte) error {
	var strInterval string
	err := json.Unmarshal(data, &strInterval)
	if err != nil {
		return err
	}
	duration, err := time.ParseDuration(strInterval)
	if err != nil {
		return err
	}
	*i = Interval(duration)
	return nil
}

func (i Interval) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(i).String())
}
