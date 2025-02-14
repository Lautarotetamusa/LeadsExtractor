package whatsapp

import (
	"bytes"
	"fmt"
	"leadsextractor/models"
	"leadsextractor/pkg/numbers"
	"text/template"
)

const (
	templateMsgAsesor = "info_asesor_2"
)

func (w *Whatsapp) SendMsgAsesor(to string, c *models.Communication, isNew bool) (*Response, error) {
	isNewMsg := "DUPLICADO"
	if isNew {
		isNewMsg = "NUEVO"
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
		isNewMsg,
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
	return w.Send(NewTemplatePayload(to, templateMsgAsesor, components))
}

func (wh *Webhook) Entry2Communication(e *Entry) (*models.Communication, error) {
	if err := wh.validateEntry(e); err != nil {
		return nil, err
	}
	//TODO: Why are multiples changes??
	value := e.Changes[0].Value

	// TODO: Read all the messages? why are multiple contacts??
	contact := value.Contacts[0]
	message := value.Messages[0]

	wh.logger.Info("new message recived",
		"phone", contact.WaID,
		"name", contact.Profile.Name,
		"id", message.Id)

	phone, err := numbers.NewPhoneNumber(contact.WaID)
	if err != nil {
		return nil, fmt.Errorf("error parsing whatsapp number: %s", contact.WaID)
	}

	messageText := wh.getMessageText(&message)

	c := models.Communication{
		Fuente:   "whatsapp",
		Nombre:   contact.Profile.Name,
		Link:     fmt.Sprintf(webSendUrl, contact.WaID),
		Telefono: *phone,
		Message:  models.NullString{String: messageText, Valid: true},
		Wamid:    models.NullString{String: message.Id, Valid: true},
	}

	return &c, nil
}

func (wh *Webhook) validateEntry(e *Entry) error {
	if len(e.Changes) <= 0 {
		return fmt.Errorf("whatsapp entry doesn't have any Changes")
	}
	value := e.Changes[0].Value

	if len(value.Statuses) > 0 {
		for _, s := range value.Statuses {
			wh.logger.Debug("message status has changed",
				"status", s.Status,
				"recipientId", s.RecipientId,
				"msgId", s.Id)
			for _, e := range s.Errors {
				wh.logger.Error(e.Message, "code", e.Code, "to", s.RecipientId)
			}
		}
		return fmt.Errorf("status changed")
	}

	if len(value.Errors) > 0 {
		for _, e := range value.Errors {
			wh.logger.Error(e.Message, "code", e.Code, "title", e.Title)
		}
		// TODO: Parse all the errors
		return fmt.Errorf("%d - %s", value.Errors[0].Code, value.Errors[0].Message)
	}

	if len(value.Contacts) <= 0 {
		return fmt.Errorf("whatsapp value doesn't have any Contacts")
	}
	if len(value.Messages) <= 0 {
		return fmt.Errorf("whatsapp value doesn't have any Messages")
	}
	return nil
}

func (wh *Webhook) getMessageText(m *Message) string {
	var text string

	switch m.Type {
	case InteractiveMessage:
		if m.Interactive != nil {
			if m.Interactive.ButtonReply != nil {
				text = m.Interactive.ButtonReply.Title
			}
		}
	case ButtonMessage:
		if m.Button != nil {
			text = m.Button.Text
		}
	case TextMessage:
		if m.Text != nil {
			text = m.Text.Body
		}
	default:
		text = ""
	}
	wh.logger.Info(fmt.Sprintf("Mensaje de tipo %s", m.Type), "text", text)
	return text
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
