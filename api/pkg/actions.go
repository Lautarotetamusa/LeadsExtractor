package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"leadsextractor/infobip"
	"leadsextractor/models"
	"leadsextractor/whatsapp"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"text/template"
	"time"

	"atomicgo.dev/schedule"
	"github.com/google/uuid"
)

type ActionFunc func(c *models.Communication, params interface{}) error

type ActionDefinition struct {
	Func        ActionFunc
	ParamType   reflect.Type
}

type SendWppTextParam struct {
    Text    string  `json:"text"`
}

type Action struct {
    ActionName  string      `json:"action"`
    Intervalo   Interval    `json:"interval"`
    Params      interface{} `json:"params"`
}

type Condition struct {
    IsNew   bool `json:"is_new"`
}

type Rule struct {
    Condition   Condition `json:"condition"`
    Actions     []Action `json:"actions"`  
}

var flows map[uuid.UUID][]Rule

var actionsDefinitions map[string]ActionDefinition

type Interval time.Duration

func (a *Action) UnmarshalJSON(data []byte) error {
    var temp struct {
        ActionName string          `json:"action"`
        Intervalo  Interval        `json:"interval"`
        Params     json.RawMessage `json:"params"`
    }

    if err := json.Unmarshal(data, &temp); err != nil {
        return err
    }

    a.ActionName = temp.ActionName
    a.Intervalo = temp.Intervalo

    actionDef, exists := actionsDefinitions[a.ActionName]
    if !exists {
        return fmt.Errorf("la accion %s no existe", a.ActionName)
    }

    //Deserealizamos al tipo de esta accion definido en ParamsType
    params := reflect.New(actionDef.ParamType).Interface()
    if err := json.Unmarshal(temp.Params, params); err != nil {
        return fmt.Errorf("parámetros inválidos para la acción %s: %v", a.ActionName, err)
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

func loadActionConfig(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    if err := json.NewDecoder(file).Decode(&flows); err != nil {
        return err
    }

    return nil
}

func saveConfigToFile(filename string) error {
    data, err := json.MarshalIndent(flows, "", "\t")
    if err != nil {
        return err
    }

    err = os.WriteFile(filename, data, 0644)
    if err != nil {
        return err
    }

    return nil
}

func ConfigureAction(w http.ResponseWriter, r *http.Request) error {
    var rules []Rule
    if err := json.NewDecoder(r.Body).Decode(&rules); err != nil {
        return err
    }

    uuid, err := uuid.NewRandom()
    if err != nil {
        return fmt.Errorf("no se pudo generar una uuid: %s", err)
    }
    flows[uuid] = rules
    saveConfigToFile("actions.json")

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(flows)
    return nil
}

func broadcast(comms []models.Communication, uuid uuid.UUID) error {
    _, ok := flows[uuid] 
    if !ok{
        return fmt.Errorf("el flow con uuid %s no existe", uuid)
    }

    for _, c := range comms {
        go runFlow(&c, uuid)
    }

    return nil
}

func runFlow(c *models.Communication, uuid uuid.UUID) {
    rules, ok := flows[uuid] 
    if !ok{
        slog.Error(fmt.Sprintf("el flow con uuid %s no existe", uuid.String()))
        os.Exit(1)
    }

    for _, rule := range rules {
        if rule.Condition.IsNew != c.IsNew {
           continue 
        }

        for _, action := range rule.Actions{
            actionFunc := actionsDefinitions[action.ActionName].Func
            
            schedule.After(time.Duration(action.Intervalo), func() {
                err := actionFunc(c, action.Params)
                if err != nil {
                    slog.Error(fmt.Sprintf("error in action %s: %s", action.ActionName, err.Error()))
                }
            })
        }
    }
}

func formatMsg(tmpl string, c *models.Communication) string{
    t := template.Must(template.New("txt").Parse(tmpl))
    buf := &bytes.Buffer{}
    if err := t.Execute(buf, c); err != nil {
        return ""
    }
    return buf.String()
}

func MustReadFile (filepath string) string {
    bytes, err := os.ReadFile(filepath)
    if err != nil {
        panic(fmt.Sprintf("error abriendo el archivo %s", filepath))
    }
    return string(bytes)
}

func (s *Server) setupActions(){
    //bienvenida1 := MustReadFile("../messages/bienvenida_1.txt")
    //bienvenida2 := MustReadFile("../messages/bienvenida_2.txt")
    cotizacion1 := MustReadFile("../messages/plantilla_cotizacion_1.txt")
    cotizacion2 := MustReadFile("../messages/plantilla_cotizacion_2.txt")

    actionsDefinitions = make(map[string]ActionDefinition)
    actionsDefinitions["wpp.message"] = ActionDefinition{
        Func: func(c *models.Communication, params interface{}) error {
            s.logger.Info("wpp.message")
            param, ok := params.(*SendWppTextParam)
            if !ok {
                return fmt.Errorf("invalid parameters for wpp.message")
            }

            msg := formatMsg(param.Text, c)
            s.whatsapp.SendMessage(c.Telefono, msg)
            return nil
        },
        ParamType: reflect.TypeOf(SendWppTextParam{}),
    }

    actionsDefinitions["wpp.cotizacion"] = ActionDefinition{
        Func: func(c *models.Communication, params interface{}) error {
            s.logger.Info("wpp.send_cotizacion")
            if c.Cotizacion == "" {
                return fmt.Errorf("el lead no tiene cotizacion")
            }

            var caption string
            if !c.Busquedas.CoveredArea.Valid {
                caption = cotizacion2
            }else{
                caption = cotizacion1
            }

            s.whatsapp.Send(whatsapp.NewDocumentPayload(
                c.Telefono,
                c.Cotizacion,
                caption,
                fmt.Sprintf("Cotizacion para %s", c.Nombre),
            ))
            return nil
        },
    }

    actionsDefinitions["wpp.send_message_asesor"] = ActionDefinition{
		Func: func(c *models.Communication, params interface{}) error {
            s.logger.Info("wpp.send_message_asesor")
            s.whatsapp.SendMsgAsesor(c.Asesor.Phone, c, c.IsNew)
            return nil
        },
    }

    actionsDefinitions["wpp.send_image"] = ActionDefinition{
		Func: func(c *models.Communication, params interface{}) error {
            s.logger.Info("wpp.send_image")
            s.whatsapp.SendImage(c.Telefono, os.Getenv("WHATSAPP_IMAGE_ID"))
            return nil
        },
    }

    actionsDefinitions["wpp.send_video"] = ActionDefinition{
		Func: func(c *models.Communication, params interface{}) error {
            s.logger.Info("wpp.send_video")
            s.whatsapp.SendVideo(c.Telefono, os.Getenv("WHATSAPP_VIDEO_ID"))
            return nil
        },
    }

    actionsDefinitions["wpp.send_response"] = ActionDefinition{
            Func: func(c *models.Communication, params interface{}) error {
            s.logger.Info("wpp_send_response")
            s.whatsapp.SendResponse(c.Telefono, &c.Asesor)
            return nil
        },
    }

    actionsDefinitions["infobip.save"] = ActionDefinition{
        Func: func(c *models.Communication, params interface{}) error {
            infobipLead := infobip.Communication2Infobip(c)
            s.infobipApi.SaveLead(infobipLead)
            return nil
        },
    }

    actionsDefinitions["pipedrive.save"] = ActionDefinition{
        Func: func(c *models.Communication, params interface{}) error {
            s.pipedrive.SaveCommunication(c)
            return nil
        },
    }
    
    err := loadActionConfig("actions.json")
    if err != nil{
        s.logger.Error(err.Error())
    }
}
