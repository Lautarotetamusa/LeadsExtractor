// Funciones relacionadas con el sistema actual
package whatsapp

import (
	"bytes"
	"fmt"
	"leadsextractor/models"
	"leadsextractor/numbers"
	"text/template"
)

const (
    template_dup_lead_asesor = "info_asesor"
    template_new_lead_asesor = "msg_asesor_duplicado"
)

func (w *Whatsapp) SendMsgAsesor(to string, c *models.Communication, isNew bool) (*Response, error) {
	templateName := template_new_lead_asesor
	if isNew {
		templateName = template_dup_lead_asesor
	}
	params := []string{
		c.Nombre,
		c.Fuente,
		c.Telefono.String(),
        c.Email.String,
		c.FechaLead,
		c.Link,
		c.Propiedad.Titulo.String,
		c.Propiedad.Precio.String,
		c.Propiedad.Ubicacion.String,
		c.Propiedad.Link.String,
        c.Busquedas.Presupuesto,
        c.Busquedas.TotalArea.String,
        c.Busquedas.CoveredArea.String,

        c.Utm.Source.String,
        c.Utm.Channel.String,
        c.Utm.Medium.String,
        c.Utm.Campaign.String,
        c.Utm.Ad.String,
	}

	parameters := make([]Parameter, len(params))
	for i, param := range params {
		if param == "" {
			param = " - "
		}
		parameters[i] = Parameter{
            Type: "text",
            Text: param,
		}
	}
	components := []Components{{
		Type:       "body",
		Parameters: parameters,
	}}
	return w.Send(NewTemplatePayload(to, templateName, components))
}

func (wh *Webhook) Entry2Communication(e *Entry) (*models.Communication, error) {
    if (len(e.Changes) <= 0) {
        return nil, fmt.Errorf("whatsapp entry doesn't have any Changes")
    }
    value := e.Changes[0].Value

    if len(value.Statuses) > 0 {
        for _, s := range value.Statuses {
            wh.logger.Debug("cambio de estado de un mensaje", 
                "status", s.Status, 
                "recipientId", s.RecipientId, 
                "msgId", s.Id)
            for _, e := range s.Errors {
                wh.logger.Error(e.Message, "code", e.Code, "to", s.RecipientId)
            }
        }
        return nil, fmt.Errorf("cambio de estado")
    }

    if len(value.Errors) > 0 {
        for _, e := range value.Errors {
            wh.logger.Error(e.Message, "code", e.Code, "title", e.Title)
        }
        return nil, fmt.Errorf("%d - %s", value.Errors[0].Code, value.Errors[0].Message)
    }

    if (len(value.Contacts) <= 0) {
        return nil, fmt.Errorf("whatsapp value doesn't have any Contacts")
    }
    if (len(value.Messages) <= 0) {
        return nil, fmt.Errorf("whatsapp value doesn't have any Messages")
    }

    // TODO: Leer todos los mensajes? por qué hay muchos contactos?
    contact := value.Contacts[0]
    message := value.Messages[0]

    wh.logger.Info("nuevo mensaje de whatsapp recibido", 
        "phone", contact.WaID, 
        "name", contact.Profile.Name, 
        "id", message.Id)
    
    phone, err := numbers.NewPhoneNumber(contact.WaID)
    if err != nil {
        return nil, fmt.Errorf("error parsing whatsapp number: %s", contact.WaID)
    }

    c := models.Communication {
        Fuente: "whatsapp",
        FechaLead: "",
        Fecha: "",
        Nombre: contact.Profile.Name,
        Link: fmt.Sprintf("https://web.whatsapp.com/send/?phone=%s", contact.WaID),
        Telefono: *phone,
        Email: models.NullString{String: ""},
        Cotizacion: "",
        Asesor: models.Asesor{},
        Propiedad: models.Propiedad{},
        Busquedas: models.Busquedas{},
        IsNew: false, // WHAT?
        Message: models.NullString{String: message.Text.Body, Valid: true},
        Wamid: models.NullString{String: message.Id, Valid: true},
    }

    return &c, nil
}

func ParseParameters(c Components, communication *models.Communication) []Parameter {
    parsedParameters := make([]Parameter, len(c.Parameters))

	for i, param := range c.Parameters {
        t, err := template.New("txt").Parse(param.Text)
        if err == nil { //Si lo puede parsear lo hace
            buf := &bytes.Buffer{}
            if err := t.Execute(buf, communication); err != nil {
                continue
            }
            parsedParam := param 
            parsedParam.Text = buf.String()

            if parsedParam.Text == "" {
                parsedParam.Text = " - "
            }

            parsedParameters[i] = parsedParam
        }
    }
    return parsedParameters
}
