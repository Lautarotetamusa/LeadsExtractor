package flow

import (
	"encoding/json"
	"fmt"
	"leadsextractor/models"
	"log/slog"
	"os"
	"time"

	"atomicgo.dev/schedule"
	"github.com/google/uuid"
)

func (f *FlowManager) MustLoad() {
    file, err := os.Open(f.filename)
    if err != nil {
        panic(err.Error())
    }
    defer file.Close()

    if err := json.NewDecoder(file).Decode(&f); err != nil {
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

func (f *FlowManager) GetFlows() map[uuid.UUID]Flow {
    return f.Flows 
}

func (f *FlowManager) GetMain() uuid.UUID {
    return f.Main 
}

func (f *FlowManager) GetFlow(uuid uuid.UUID) (*Flow, error){
    flow, ok := f.Flows[uuid] 
    if !ok{
        return nil, fmt.Errorf("el flow con uuid %s no existe", uuid)
    }
    return &flow, nil
}

func (f *FlowManager) GetMainFlow() (*Flow, error){
    flow, ok := f.Flows[f.Main] 
    if !ok{
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

func (f *FlowManager) AddFlow(flow *Flow) error {
    uuid, err := uuid.NewRandom()
    if err != nil {
        return fmt.Errorf("no se pudo generar una uuid: %s", err)
    }
    flow.Uuid = uuid

    f.Flows[uuid] = *flow
    if err := f.Save(); err != nil {
        return err
    }
    
    return nil
}

func (f *FlowManager) UpdateFlow(flow *Flow, uuid uuid.UUID) error {
    f.Flows[uuid] = *flow
    if err := f.Save(); err != nil {
        return err
    }
    
    return nil
}

func (f *FlowManager) DeleteFlow(uuid uuid.UUID) error {
    if uuid == f.Main {
        return fmt.Errorf("no se puede eliminar el flow principal")
    }

    delete(f.Flows, uuid)

    if err := f.Save(); err != nil {
        return err
    }

    return nil
}

func (f * FlowManager) Broadcast(comms []models.Communication, uuid uuid.UUID) error {
    f.logger.Info("lanzando broadcast", "count", len(comms))
    _, ok := f.Flows[uuid] 
    if !ok{
        return fmt.Errorf("el flow con uuid %s no existe", uuid)
    }

    for _, c := range comms {
        go f.runFlow(&c, uuid)
        time.Sleep(100 * time.Millisecond)
    }
    return nil
}

func (f *FlowManager) RunMainFlow(c *models.Communication) {
    f.runFlow(c, f.Main)
}

func (f *FlowManager) runFlow(c *models.Communication, uuid uuid.UUID) {
    flow, ok := f.Flows[uuid] 
    if !ok{
        slog.Error(fmt.Sprintf("el flow con uuid %s no existe", uuid.String()))
        os.Exit(1)
    }

    for _, rule := range flow.Rules {
        if !rule.Condition.Matches(c) {
            continue
        }

        for _, action := range rule.Actions{
            actionFunc := actions[action.Name].Func
            
            schedule.After(time.Duration(action.Interval), func() {
                f.logger.Debug("running action", "name", action.Name)
                err := actionFunc(c, action.Params)
                if err != nil {
                    f.logger.Error(err.Error(), "action", action.Name)
                }
            })
        }
    }
}
