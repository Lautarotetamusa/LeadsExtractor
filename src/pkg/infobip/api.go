package infobip

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"leadsextractor/models"
	"log"
	"log/slog"
	"net/http"
	"net/url"
)

func NewInfobipApi(apiUrl string, apiKey string, sender string, l *slog.Logger) *InfobipApi {
	return &InfobipApi{
		apiUrl: apiUrl,
		apiKey: apiKey,
		sender: sender,
		client: &http.Client{},
		logger: l.With("module", "infobip"),
	}
}

func (i *InfobipApi) SendWppTemplate(to string, templateName string, language string) {
	if language == "" {
		language = "es_MX"
	}

	message := SendMessage{
		From: i.sender,
		To:   to,
		Content: TemplateContent{
			TemplateName: templateName,
			Language:     language,
		},
	}
	payload := &SendTemplatesPayload{
		Messages: []SendMessage{message},
	}

	go i.makeRequest("POST", "/whatsapp/1/message/template", payload)
}

func (i *InfobipApi) SendWppMediaMessage(to string, tipo string, mediaUrl string) {
	payload := &SendMessage{
		From: i.sender,
		To:   to,
		Content: MediaContent{
			MediaUrl: mediaUrl,
		},
	}

	url := fmt.Sprintf("/whatsapp/1/message/%s", tipo)
	go i.makeRequest("POST", url, payload)
}

func (i *InfobipApi) SendWppTextMessage(to string, message string) {
	log.Println("Enviando mensaje de texto por whatsapp")
	payload := &SendMessage{
		From: i.sender,
		To:   to,
		Content: TextContent{
			Text: message,
		},
	}

	go i.makeRequest("POST", "/whatsapp/1/message/text", payload)
}

func (i *InfobipApi) SaveLead(lead *InfobipLead) {
	i.logger.Info("Cargando lead")
	go i.makeRequest("POST", "/people/2/persons", lead)
}

func Communication2Infobip(c *models.Communication) *InfobipLead {
	var contactInformation ContactInformation
	if c.Email.Valid {
		contactInformation.Email = &EmailContact{
			Address: c.Email.String,
		}
	}
	if c.Telefono != "" {
		contactInformation.Phone = &PhoneContact{
			Number: c.Telefono,
		}
	}

	return &InfobipLead{
		FirstName: c.Nombre,
		LastName:  "",
		Type:      "LEAD",
		CustomAttributes: CustomAttributes{
			PropLink:      c.Propiedad.Link.String,
			PropPrecio:    c.Propiedad.Precio.String,
			PropUbicacion: c.Propiedad.Ubicacion.String,
			PropTitulo:    c.Propiedad.Titulo.String,
			Contacted:     false,
			Fuente:        c.Fuente,
			AsesorName:    c.Asesor.Name,
			AsesorPhone:   c.Asesor.Phone.String(),
			FechaLead:     c.FechaLead,
		},
		ContactInformation: contactInformation,
		Tags:               "Seguimientos",
	}
}

// minBalance es el umbral por debajo del cual se avisa que la cuenta
// prepaga se está por quedar sin saldo.
const minBalance = 10.0

type accountBalance struct {
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
}

func (i *InfobipApi) Name() string {
	return "infobip"
}

func (i *InfobipApi) getBalance(ctx context.Context) (*accountBalance, error) {
	balanceUrl, err := url.JoinPath(i.apiUrl, "/account/1/balance")
	if err != nil {
		return nil, fmt.Errorf("la url es incorrecta: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, balanceUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", i.apiKey)

	res, err := i.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("no se pudo realizar la peticion: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("respuesta inesperada de infobip: %d %s", res.StatusCode, body)
	}

	var balance accountBalance
	if err := json.Unmarshal(body, &balance); err != nil {
		return nil, fmt.Errorf("no se pudo parsear la respuesta de infobip: %w", err)
	}
	return &balance, nil
}

// HealthCheck valida el API key consultando el balance de la cuenta, una
// llamada de solo lectura que no envía ningún mensaje.
func (i *InfobipApi) HealthCheck(ctx context.Context) error {
	_, err := i.getBalance(ctx)
	return err
}

// UsageWarning avisa cuando el saldo prepago de la cuenta cae por debajo de
// minBalance, para enterarse antes de que Infobip empiece a rechazar envíos
// por falta de saldo.
func (i *InfobipApi) UsageWarning(ctx context.Context) (string, error) {
	balance, err := i.getBalance(ctx)
	if err != nil {
		return "", err
	}

	if balance.Balance < minBalance {
		return fmt.Sprintf("saldo bajo: %.2f %s", balance.Balance, balance.Currency), nil
	}
	return "", nil
}

func (i *InfobipApi) makeRequest(method string, path string, payload interface{}) {
	postBody, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		i.logger.Error("No se pudo parsear el payload", "err", err)
		return
	}
	data := bytes.NewBuffer(postBody)

	url, err := url.JoinPath(i.apiUrl, path)
	if err != nil {
		i.logger.Error("La url es incorrecta", "path", path)
		return
	}

	req, err := http.NewRequest(method, url, data)
	if err != nil {
		i.logger.Error("No se pudo construir la request", "err", err)
		return
	}

	req.Header.Add("Authorization", i.apiKey)
	req.Header.Add("Content-Type", "application/json")

	res, err := i.client.Do(req)
	if err != nil {
		i.logger.Error("Error en la request", "err", err)
		return
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		i.logger.Error("Error obteniendo la respuesta", "err", err)
		return
	}

	if res.StatusCode != http.StatusOK {
		i.logger.Warn("Response not ok", "status", res.StatusCode, "res", string(body))
		return
	}

	i.logger.Info("Request enviada correctamente", "res", string(body))
}
