package flow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"leadsextractor/models"
	"leadsextractor/pkg/email"
	"leadsextractor/pkg/infobip"
	"leadsextractor/pkg/jotform"
	"leadsextractor/pkg/pipedrive"
	"leadsextractor/pkg/whatsapp"
	"leadsextractor/store"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"text/template"
)

// Definimos las acciones permitidas dentro de un flow
func DefineActions(
	wpp *whatsapp.Whatsapp,
	pipedriveApi *pipedrive.Pipedrive,
	infobipApi *infobip.InfobipApi,
	leadStorer store.LeadStorer,
	mailer email.Sender,
) {
	cotizacion1 := mustReadFile("../messages/plantilla_cotizacion_1.txt")
	cotizacion2 := mustReadFile("../messages/plantilla_cotizacion_2.txt")
	jotformApi := jotform.NewJotform(
		os.Getenv("JOTFORM_API_KEY"),
		os.Getenv("APP_HOST"),
	)
	form := jotformApi.AddForm(os.Getenv("JOTFORM_FORM_ID"))

	DefineAction("wpp.message",
		func(c *models.Communication, params interface{}) error {
			param, ok := params.(*SendWppTextParam)
			if !ok {
				return fmt.Errorf("invalid parameters for wpp.message")
			}

			msg := formatMsg(param.Text, c)
			wpp.SendMessage(c.Telefono.String(), msg)
			c.LastSentMessage = models.NullString{String: msg, Valid: true}
			return nil
		},
		reflect.TypeOf(SendWppTextParam{}),
	)

	DefineAction("wpp.template",
		func(c *models.Communication, params interface{}) error {
			payload, ok := params.(*whatsapp.TemplatePayload)
			if !ok {
				return fmt.Errorf("invalid parameters for wpp.template")
			}

			// Hacemos un deep copy para no alterar los parametros
			parsedPayload := *payload
			parsedPayload.Components = make([]whatsapp.Components, len(payload.Components))
			for i, comp := range payload.Components {
				parsedPayload.Components[i] = comp
				parsedParameters := whatsapp.ParseParameters(comp, c)
				parsedPayload.Components[i].Parameters = parsedParameters
			}
			wpp.SendTemplate(c.Telefono.String(), parsedPayload)
			return nil
		},
		reflect.TypeOf(whatsapp.TemplatePayload{}),
	)

	DefineAction("wpp.media",
		func(c *models.Communication, params interface{}) error {
			payload, ok := params.(*SendWppMedia)
			if !ok {
				return fmt.Errorf("invalid parameters for wpp.media")
			}

			if payload.Image != nil && payload.Video == nil {
				wpp.SendImage(c.Telefono.String(), payload.Image.Id)
			} else if payload.Video != nil && payload.Image == nil {
				wpp.SendVideo(c.Telefono.String(), payload.Video.Id)
			} else {
				return fmt.Errorf("parameters 'image' or 'video' not found in wpp.media")
			}
			return nil
		},
		reflect.TypeOf(SendWppMedia{}),
	)

	DefineAction("wpp.cotizacion",
		func(c *models.Communication, params interface{}) error {
			if c.Cotizacion == "" {
				url, err := jotformApi.GetPdf(c, form)
				if err != nil {
					return err
				}
				c.Cotizacion = url
				lead := models.Lead{
					Name:       c.Nombre,
					Email:      c.Email,
					Phone:      c.Telefono,
					Cotizacion: c.Cotizacion,
				}
				if err := leadStorer.Update(&lead); err != nil {
					return err
				}
				slog.Info("Cotizacion generada con exito")
			}

			var caption string
			if !c.Busquedas.CoveredArea.Valid {
				caption = cotizacion2
			} else {
				caption = cotizacion1
			}

			wpp.Send(whatsapp.NewDocumentPayload(
				c.Telefono.String(),
				c.Cotizacion,
				caption,
				fmt.Sprintf("Cotizacion para %s", c.Nombre),
			))
			return nil
		},
		nil,
	)

	DefineAction("wpp.send_message_asesor",
		func(c *models.Communication, params interface{}) error {
			wpp.SendMsgAsesor(c.Asesor.Phone.String(), c, c.IsNew)
			return nil
		},
		nil,
	)

	DefineAction("infobip.save",
		func(c *models.Communication, params interface{}) error {
			infobipLead := infobip.Communication2Infobip(c)
			infobipApi.SaveLead(infobipLead)
			return nil
		},
		nil,
	)

	DefineAction("pipedrive.save",
		func(c *models.Communication, params interface{}) error {
			pipedriveApi.SaveCommunication(c)
			return nil
		},
		nil,
	)

	DefineAction("email.send_asesor",
		func(c *models.Communication, params interface{}) error {
			if c.Asesor.Email == "" {
				return fmt.Errorf("el asesor %s no tiene email configurado", c.Asesor.Name)
			}
			isNew := "DUPLICADO"
			if c.IsNew {
				isNew = "NUEVO"
			}
			subject := fmt.Sprintf("[%s] Lead de %s - %s", isNew, c.Fuente, c.Nombre)
			body := buildAsesorEmailHTML(c)
			return mailer.Send(
				context.Background(), 
				"from@rbaresidences.com",
				[]string{c.Asesor.Email}, 
				subject, 
				body,
				true,
			)
		},
		nil,
	)
}

func (m *FlowManager) GetActions() interface{} {
	file, err := os.Open("flow/config.json")
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	var config interface{}

	if err := json.NewDecoder(file).Decode(&config); err != nil {
		panic(fmt.Sprintf("cannot load config.json: %s", err.Error()))
	}

	return config
}

func mustReadFile(filepath string) string {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		panic(fmt.Sprintf("error abriendo el archivo %s", filepath))
	}
	return string(bytes)
}

func formatMsg(tmpl string, c *models.Communication) string {
	t := template.Must(template.New("txt").Parse(tmpl))
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, c); err != nil {
		return ""
	}
	return buf.String()
}

func buildAsesorEmailHTML(c *models.Communication) string {
	isNew := "DUPLICADO"
	if c.IsNew {
		isNew = "NUEVO"
	}

	rows := [][2]string{
		{"Estado", isNew},
		{"Fuente", c.Fuente},
		{"Nombre", c.Nombre},
		{"Teléfono", c.Telefono.String()},
		{"Email", c.Email.String},
		{"Fecha", c.FechaLead},
		{"Link", c.Link},
		{"Propiedad", c.Propiedad.Titulo.String},
		{"Precio", c.Propiedad.Precio.String},
		{"Ubicación", c.Propiedad.Ubicacion.String},
		{"Link propiedad", c.Propiedad.Link.String},
		{"Presupuesto", c.Busquedas.Presupuesto},
		{"UTM source", c.Utm.Source.String},
		{"UTM campaign", c.Utm.Campaign.String},
	}

	var sb strings.Builder
	sb.WriteString(`<html><body style="font-family:sans-serif">`)
	sb.WriteString(fmt.Sprintf(`<h2>Nuevo lead: %s</h2>`, c.Nombre))
	sb.WriteString(`<table border="1" cellpadding="6" cellspacing="0" style="border-collapse:collapse">`)

	for _, row := range rows {
		if row[1] == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf(
			`<tr><td style="background:#f0f0f0;font-weight:bold">%s</td><td>%s</td></tr>`,
			row[0], row[1],
		))
	}

	sb.WriteString(`</table></body></html>`)
	return sb.String()
}
