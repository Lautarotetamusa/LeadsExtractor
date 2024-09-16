package flow

import (
	"encoding/json"
	"fmt"
	"leadsextractor/models"
	"leadsextractor/store"
	"leadsextractor/whatsapp"
	"log/slog"
	"reflect"
	"time"

	"github.com/google/uuid"
)

type Interval time.Duration

type ActionFunc func(c *models.Communication, params interface{}) error

type ActionDefinition struct {
	Func        ActionFunc
	ParamType   reflect.Type
}

type SendWppTextParam struct {
    Text string  `json:"text"`
}

type SendWppMedia struct {
    Image   *whatsapp.MediaPayload    `json:"image,omitempty" jsonschema:"oneof_required=image"`
    Video   *whatsapp.MediaPayload    `json:"video,omitempty" jsonschema:"oneof_required=video"`
}

type Action struct {
    Name        string      `json:"action"`
    Interval    Interval    `json:"interval"`
    Params      interface{} `json:"params"`
}

type Rule struct {
    Condition   store.QueryParam    `json:"condition"`
    Actions     []Action            `json:"actions"`  
}

type Flow struct {
    Rules   []Rule      `json:"rules"`
    Name    string      `json:"name"`
    Uuid    uuid.UUID   `json:"uuid"`
    IsMain  bool        `json:"is_main"`
}

type FlowManager struct {
    Main        uuid.UUID           //Es el uuid del flow que se ejecuta cuando llega una comuncacion
    Flows       map[uuid.UUID]Flow 
    filename    string
    logger      *slog.Logger
}

//Son las acciones permitidas
var actions map[string]ActionDefinition

func DefineAction(name string, f ActionFunc, t reflect.Type) {
    actions[name] = ActionDefinition{
        Func: f,
        ParamType: t,
    }
}

func NewFlowManager(filename string, l *slog.Logger) *FlowManager {
    actions = make(map[string]ActionDefinition)
    return &FlowManager{
        filename: filename,
        logger: l.With("module", "flow"),
        Flows: make(map[uuid.UUID]Flow),
    }
}

func (a *Action) UnmarshalJSON(data []byte) error {
    var temp struct {
        Name        string          `json:"action"`
        Interval    Interval        `json:"interval"`
        Params      json.RawMessage `json:"params"`
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
