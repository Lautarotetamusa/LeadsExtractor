package flow

import (
	"encoding/json"
	"fmt"
	"leadsextractor/models"
	"leadsextractor/store"
	"os"
	"time"

	"atomicgo.dev/schedule"
	"github.com/google/uuid"
)

type flowManager interface {
	GetAll() map[uuid.UUID]Flow
	GetMainUUID() uuid.UUID
	GetOne(uuid uuid.UUID) (*Flow, error)
	GetMain() (*Flow, error)

	SetMain(uuid uuid.UUID) error
	Add(flow *Flow) (*uuid.UUID, error)
	Update(flow *Flow, uuid uuid.UUID) error
	Delete(uuid uuid.UUID) error

	Broadcast(comms []models.Communication, uuid uuid.UUID) error
	RunMain(c *models.Communication)
	Run(c *models.Communication, uuid uuid.UUID)
}

func (f *FlowManager) MustLoad() {
	file, err := os.Open(f.filename)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&f); err != nil {
		fmt.Printf("%#v\n", err)
		panic(fmt.Sprintf("cannot load actions.json: %s", err.Error()))
	}
}

func (f *FlowManager) Save() error {
	data, err := json.MarshalIndent(f, "", "\t")
	if err != nil {
		return err
	}

	if err = os.WriteFile(f.filename, data, 0644); err != nil {
		return err
	}

	return nil
}

func (f *FlowManager) GetAll() map[uuid.UUID]Flow {
	return f.Flows
}

func (f *FlowManager) GetMainUUID() uuid.UUID {
	return f.Main
}

func (f *FlowManager) GetOne(uuid uuid.UUID) (*Flow, error) {
	flow, ok := f.Flows[uuid]
	if !ok {
		return nil, fmt.Errorf("el flow con uuid %s no existe", uuid)
	}
	return &flow, nil
}

func (f *FlowManager) GetMain() (*Flow, error) {
	flow, ok := f.Flows[f.Main]
	if !ok {
		return nil, fmt.Errorf("no hay ningun flow main")
	}
	return &flow, nil
}

func (f *FlowManager) SetMain(uuid uuid.UUID) error {
	f.Main = uuid

	if err := f.Save(); err != nil {
		return err
	}
	return nil
}

func (f *FlowManager) Add(flow *Flow) (*uuid.UUID, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("no se pudo generar una uuid: %s", err)
	}

	if err := f.validateOnResponse(flow); err != nil {
		return nil, err
	}

	f.Flows[uuid] = *flow
	if err := f.Save(); err != nil {
		return nil, err
	}

	return &uuid, nil
}

func (f *FlowManager) Update(flow *Flow, uuid uuid.UUID) error {
	f.Flows[uuid] = *flow
	if err := f.Save(); err != nil {
		return err
	}

	return nil
}

func (f *FlowManager) Delete(uuid uuid.UUID) error {
	if uuid == f.Main {
		return fmt.Errorf("no se puede eliminar el flow principal")
	}

	delete(f.Flows, uuid)

	if err := f.Save(); err != nil {
		return err
	}

	return nil
}

func (f *FlowManager) Broadcast(comms []models.Communication, uuid uuid.UUID) error {
	f.logger.Info("Lanzando broadcast", "count", len(comms))
	_, ok := f.Flows[uuid]
	if !ok {
		return fmt.Errorf("el flow con uuid %s no existe", uuid)
	}

	for _, c := range comms {
		go f.RunFlow(&c, uuid)
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (f *FlowManager) RunMain(c *models.Communication) {
	f.RunFlow(c, f.Main)
}

func (f *FlowManager) RunFlow(c *models.Communication, uuid uuid.UUID) {
	flow, ok := f.Flows[uuid]
	if !ok {
		f.logger.Error(fmt.Sprintf("el flow con uuid %s no existe", uuid.String()))
		os.Exit(1)
	}
	f.logger.Debug("running flow", "uuid", uuid)

	for _, rule := range flow.Rules {
		if !rule.Condition.Matches(c) {
			continue
		}

		for order, action := range rule.Actions {
			actionFunc := actions[action.Name].Func
			actionRunned := &store.ActionSave{
				Name:       action.Name,
				Nro:        order,
				FlowUUID:   uuid,
				LeadPhone:  c.Telefono.String(),
				OnResponse: rule.OnResponse,
			}

			schedule.After(time.Duration(action.Interval), func() {
				f.logger.Debug("running action", "name", action.Name)

				if err := f.storer.SaveAction(actionRunned); err != nil {
					f.logger.Error("cannot save action", "err", err.Error())
				}

				err := actionFunc(c, action.Params)
				if err != nil {
					f.logger.Error(err.Error(), "action", action.Name)
				}
			})
		}
	}
}

func (f *FlowManager) validateOnResponse(flow *Flow) error {
	for _, rule := range flow.Rules {
		if !rule.OnResponse.Valid {
			continue
		}
		_, ok := f.Flows[rule.OnResponse.UUID]
		if !ok {
			return fmt.Errorf("on_respose=%s no corresponde a ningun flow", rule.OnResponse.UUID)
		}
	}
	return nil
}
