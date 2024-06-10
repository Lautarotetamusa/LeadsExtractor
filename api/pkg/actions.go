package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"leadsextractor/infobip"
	"leadsextractor/models"
	"leadsextractor/whatsapp"
	"log/slog"
	"os"
	"reflect"
	"text/template"
	"time"

	"atomicgo.dev/schedule"
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
type SendWppTemplateParam struct {
    Template whatsapp.TemplatePayload `json:"template"`
}

type Action struct {
    Name        string      `json:"action"`
    Interval    Interval    `json:"interval"`
    Params      interface{} `json:"params"`
}

type Condition struct {
    IsNew   bool `json:"is_new"`
}

type Rule struct {
    Condition   Condition `json:"condition"`
    Actions     []Action `json:"actions"`  
}

type Config struct {
    Flows   map[uuid.UUID][]Rule    `json:"flows"` // 
    Main    uuid.UUID               `json:"main"`  //Es el uuid del flow que se ejecuta cuando llega una comuncacion
}

var config Config

//Son las acciones permitidas
var actions map[string]ActionDefinition

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

func loadConfig(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    if err := json.NewDecoder(file).Decode(&config); err != nil {
        return err
    }

    return nil
}

func saveConfig(filename string) error {
    data, err := json.MarshalIndent(config, "", "\t")
    if err != nil {
        return err
    }

    err = os.WriteFile(filename, data, 0644)
    if err != nil {
        return err
    }

    return nil
}

func broadcast(comms []models.Communication, uuid uuid.UUID) error {
    _, ok := config.Flows[uuid] 
    if !ok{
        return fmt.Errorf("el flow con uuid %s no existe", uuid)
    }

    for _, c := range comms {
        go runFlow(&c, uuid)
    }

    return nil
}

func runMainFlow(c *models.Communication) {
    runFlow(c, config.Main)
}

func runFlow(c *models.Communication, uuid uuid.UUID) {
    rules, ok := config.Flows[uuid] 
    if !ok{
        slog.Error(fmt.Sprintf("el flow con uuid %s no existe", uuid.String()))
        os.Exit(1)
    }

    for _, rule := range rules {
        if rule.Condition.IsNew != c.IsNew {
           continue 
        }

        for _, action := range rule.Actions{
            actionFunc := actions[action.Name].Func
            
            schedule.After(time.Duration(action.Interval), func() {
                err := actionFunc(c, action.Params)
                if err != nil {
                    slog.Error(fmt.Sprintf("error in action %s: %s", action.Name, err.Error()))
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

    actions = make(map[string]ActionDefinition)
    actions["wpp.message"] = ActionDefinition{
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

    actions["wpp.template"] = ActionDefinition{
        Func:  func(c *models.Communication, params interface{}) error {
            s.logger.Info("wpp.template")
            param, ok := params.(*whatsapp.TemplatePayload)
            if !ok {
                return fmt.Errorf("invalid parameters for wpp.message")
            }

            whatsapp.ParseTemplatePayload(param)
            s.whatsapp.SendTemplate(c.Telefono, *param)
            return nil
        },
        ParamType: reflect.TypeOf(whatsapp.TemplatePayload{}), 
    }

    actions["wpp.cotizacion"] = ActionDefinition{
        Func: func(c *models.Communication, params interface{}) error {
            s.logger.Info("wpp.send_cotizacion")
            if c.Cotizacion == "" {
                return fmt.Errorf("el lead %s no tiene cotizacion", c.Nombre)
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

    actions["wpp.send_message_asesor"] = ActionDefinition{
		Func: func(c *models.Communication, params interface{}) error {
            s.logger.Info("wpp.send_message_asesor")
            s.whatsapp.SendMsgAsesor(c.Asesor.Phone, c, c.IsNew)
            return nil
        },
    }

    actions["wpp.send_image"] = ActionDefinition{
		Func: func(c *models.Communication, params interface{}) error {
            s.logger.Info("wpp.send_image")
            s.whatsapp.SendImage(c.Telefono, os.Getenv("WHATSAPP_IMAGE_ID"))
            return nil
        },
    }

    actions["wpp.send_video"] = ActionDefinition{
		Func: func(c *models.Communication, params interface{}) error {
            s.logger.Info("wpp.send_video")
            s.whatsapp.SendVideo(c.Telefono, os.Getenv("WHATSAPP_VIDEO_ID"))
            return nil
        },
    }

    actions["wpp.send_response"] = ActionDefinition{
            Func: func(c *models.Communication, params interface{}) error {
            s.logger.Info("wpp_send_response")
            s.whatsapp.SendResponse(c.Telefono, &c.Asesor)
            return nil
        },
    }

    actions["infobip.save"] = ActionDefinition{
        Func: func(c *models.Communication, params interface{}) error {
            infobipLead := infobip.Communication2Infobip(c)
            s.infobipApi.SaveLead(infobipLead)
            return nil
        },
    }

    actions["pipedrive.save"] = ActionDefinition{
        Func: func(c *models.Communication, params interface{}) error {
            s.pipedrive.SaveCommunication(c)
            return nil
        },
    }
    
    err := loadConfig("actions.json")
    if err != nil{
        s.logger.Error(err.Error())
    }
}
