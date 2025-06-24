package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const baseUrl = "https://graph.facebook.com/v17.0/%s/messages"
const webSendUrl = "https://web.whatsapp.com/send/?phone=%s"

type Whatsapp struct {
	accessToken string
	numberId    string
	client      *http.Client
	url         string
	logger      *slog.Logger
}

type Response struct {
	Contacts []struct {
		Input string `json:"input"`
		WaId  string `json:"wa_id"`
	}
	Messages []struct {
		Id string `json:"id"`
	}
}

type TextPayload struct {
	PreviewUrl bool   `json:"preview_url"`
	Body       string `json:"body"`
}

type MediaPayload struct {
	Id string `json:"id"`
}

type DocumentPayload struct {
	Link     string `json:"link"`
	Caption  string `json:"caption"`
	Filename string `json:"filename"`
}

type TemplatePayload struct {
	Name       string       `json:"name"`
	Language   Language     `json:"language"`
	Components []Components `json:"components"`
}

type Parameter struct {
	Type     string           `json:"type" jsonschema:"enum=text,enum=image,enum=video,enum=document"`
	Text     string           `json:"text,omitempty" jsonschema:"oneof_required=text"`
	Image    *MediaPayload    `json:"image,omitempty" jsonschema:"oneof_required=image"`
	Video    *MediaPayload    `json:"video,omitempty" jsonschema:"oneof_required=video"`
	Document *DocumentPayload `json:"document,omitempty" jsonschema:"oneof_required=document"`
}

type Components struct {
	Type       string      `json:"type" jsonschema:"enum=body,enum=header"`
	Parameters []Parameter `json:"parameters"`
}

type Language struct {
	Code string `default:"es_MX" json:"code" jsonschema:"enum=es_MX,enum=es"`
}

type Payload struct {
	MessagingProduct string           `default:"whatsapp" json:"messaging_product"`
	RecipientType    string           `default:"individual" json:"recipient_type"`
	To               string           `json:"to"`
	Type             string           `json:"type"`
	Text             *TextPayload     `json:"text,omitempty"`
	Template         *TemplatePayload `json:"template,omitempty"`
	Image            *MediaPayload    `json:"image,omitempty"`
	Video            *MediaPayload    `json:"video,omitempty"`
	Document         *DocumentPayload `json:"document,omitempty"`
}

func NewWhatsapp(accesToken string, numberId string, l *slog.Logger) *Whatsapp {
	w := &Whatsapp{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		accessToken: accesToken,
		numberId:    numberId,
		url:         fmt.Sprintf(baseUrl, numberId),
		logger:      l.With("module", "whatsapp"),
	}
	return w
}

// TODO: use MessageType
func newPayload(to string, tipo string) *Payload {
	return &Payload{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               to,
		Type:             tipo,
	}
}

func newTextPayload(to string, message string) *Payload {
	p := newPayload(to, "text")
	p.Text = &TextPayload{
		PreviewUrl: false,
		Body:       message,
	}
	return p
}

func NewTemplatePayload(to string, name string, components []Components) *Payload {
	p := newPayload(to, "template")
	p.Template = &TemplatePayload{
		Name:       name,
		Components: components,
		Language: Language{
			Code: "es_MX",
		},
	}
	return p
}

func NewMediaPayload(to string, id string, tipo string) *Payload {
	if tipo != "image" && tipo != "video" {
		panic("El tipo de media tiene que ser o 'video' o 'image")
	}
	p := newPayload(to, tipo)

	m := &MediaPayload{
		Id: id,
	}
	if tipo == "image" {
		p.Image = m
	} else {
		p.Video = m
	}
	return p
}

func NewDocumentPayload(to string, link string, caption string, filename string) *Payload {
	p := newPayload(to, "document")
	p.Document = &DocumentPayload{
		Link:     link,
		Caption:  caption,
		Filename: filename,
	}
	return p
}

func (w *Whatsapp) Send(payload *Payload) (*Response, error) {
	jsonBody, _ := json.Marshal(payload)
	w.logger.Debug("Enviando mensaje", "to", payload.To, "type", payload.Type)
	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodPost, w.url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", w.accessToken))

	res, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("no se pudo realizar la peticion: %s", err)
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error al leer el cuerpo de la respuesta: %w", err)
	}

	var data Response
	err = json.Unmarshal(bodyBytes, &data)
	if len(data.Messages) == 0 {
		// Debugging the payload
		j, _ := json.MarshalIndent(payload, " ", "\t")
		fmt.Println(string(j))

		// Debugging the response
		var debug interface{}
		_ = json.Unmarshal(bodyBytes, &debug)
		w.logger.Error(fmt.Sprintf("response: %v", debug))
		return nil, fmt.Errorf("no se pudo obtener el json de la peticion: %w", err)
	}

	w.logger.Info("Mensaje enviando correctamente", "to", payload.To, "type", payload.Type, "id", data.Messages[0].Id)
	return &data, nil
}

func (w *Whatsapp) SendTemplate(to string, t TemplatePayload) (*Response, error) {
	p := newPayload(to, "template")
	p.Template = &t
	return w.Send(p)
}

func (w *Whatsapp) SendMessage(to string, message string) (*Response, error) {
	return w.Send(newTextPayload(to, message))
}

func (w *Whatsapp) SendImage(to string, imageId string) {
	w.Send(NewMediaPayload(to, imageId, "image"))
}

func (w *Whatsapp) SendVideo(to string, videoId string) {
	w.Send(NewMediaPayload(to, videoId, "video"))
}
