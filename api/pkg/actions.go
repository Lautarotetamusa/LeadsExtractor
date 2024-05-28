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
	"text/template"
	"time"

	"atomicgo.dev/schedule"
)

type ActionFunc func(c *models.Communication) error

type Action struct {
    ActionName  string     `json:"action"`
    Intervalo   Interval   `json:"interval"`
}

type Condition struct {
    IsNew   bool `json:"is_new"`
}

type Rule struct {
    Condition   Condition `json:"condition"`
    Actions     []Action `json:"actions"`  
}

var rules []Rule

var actions map[string]ActionFunc

type Interval time.Duration

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

    if err := json.NewDecoder(file).Decode(&rules); err != nil {
        return err
    }

    return nil
}

func saveConfigToFile(filename string) error {
    data, err := json.MarshalIndent(rules, "", "\t")
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
    if err := json.NewDecoder(r.Body).Decode(&rules); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return nil
    }

    for _, rule := range rules {
        for _, action := range rule.Actions {
            if _, exists := actions[action.ActionName]; !exists {
                return fmt.Errorf("la accion %s no existe", action.ActionName)
            }
        }
    }

    saveConfigToFile("actions.json")

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(rules)
    return nil
}

func runActions(c *models.Communication) {
    fmt.Printf("%v\n", rules)
    for _, rule := range rules {
        if rule.Condition.IsNew != c.IsNew {
           continue 
        }

        for _, action := range rule.Actions{
            actionFunc, exists := actions[action.ActionName]

            if !exists {
                slog.Warn(fmt.Sprintf("no se encontró el acción: %s", action.ActionName))
                continue
            }

            schedule.After(time.Duration(action.Intervalo), func() {
                actionFunc(c)
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
    loadActionConfig("actions.json")

    bienvenida1 := MustReadFile("../messages/bienvenida_1.txt")
    bienvenida2 := MustReadFile("../messages/bienvenida_2.txt")
    cotizacion1 := MustReadFile("../messages/plantilla_cotizacion_1.txt")
    cotizacion2 := MustReadFile("../messages/plantilla_cotizacion_2.txt")

    actions = make(map[string]ActionFunc)
    actions["wpp.bienvenida_1"] = func(c *models.Communication) error {
        s.logger.Info("wpp.bienvenida1")
        s.whatsapp.SendMessage(c.Telefono, string(bienvenida1))
        return nil
    }

    actions["wpp.bienvenida_2"] = func(c *models.Communication) error {
        s.logger.Info("wpp.bienvenida2")
        msg := formatMsg(string(bienvenida2), c)
        s.whatsapp.SendMessage(c.Telefono, msg)
        return nil
    }

    actions["wpp.cotizacion"] = func(c *models.Communication) error {
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
    }

    actions["wpp.send_message_asesor"] = func(c *models.Communication) error {
        s.logger.Info("wpp.send_message_asesor")
        s.whatsapp.SendMsgAsesor(c.Asesor.Phone, c, c.IsNew)
        return nil
    }

    actions["wpp.send_image"] = func(c *models.Communication) error {
        s.logger.Info("wpp.send_image")
        s.whatsapp.SendImage(c.Telefono, os.Getenv("WHATSAPP_IMAGE_ID"))
        return nil
    }

    actions["wpp.send_video"] = func(c *models.Communication) error {
        s.logger.Info("wpp.send_video")
        s.whatsapp.SendVideo(c.Telefono, os.Getenv("WHATSAPP_VIDEO_ID"))
        return nil
    }

    actions["wpp.send_response"] = func(c *models.Communication) error {
        s.logger.Info("wpp_send_response")
        s.whatsapp.SendResponse(c.Telefono, &c.Asesor)
        return nil
    }

    actions["infobip.save"] = func(c *models.Communication) error {
        infobipLead := infobip.Communication2Infobip(c)
        s.infobipApi.SaveLead(infobipLead)
        return nil
    }

    actions["pipedrive.save"] = func(c *models.Communication) error {
	    s.pipedrive.SaveCommunication(c)
        return nil
    }
}
