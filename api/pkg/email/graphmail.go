package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Config contiene los parámetros de tu aplicación registrada en Azure AD
type Config struct {
	ClientID     string
	TenantID     string
	ClientSecret string
}

// Message representa la estructura de un correo para Graph API
type Message struct {
	Subject         string        `json:"subject"`
	Body            MessageBody   `json:"body"`
	ToRecipients    []Recipient   `json:"toRecipients"`
	CcRecipients    []Recipient   `json:"ccRecipients,omitempty"`
	BccRecipients   []Recipient   `json:"bccRecipients,omitempty"`
	From            *Recipient    `json:"from,omitempty"` // opcional, el remitente se define en la URL
}

type MessageBody struct {
	ContentType string `json:"contentType"` // "HTML" o "Text"
	Content     string `json:"content"`
}

type Recipient struct {
	EmailAddress EmailAddress `json:"emailAddress"`
}

type EmailAddress struct {
	Address string `json:"address"`
}

// GraphMailer es el cliente que envía correos
type GraphMailer struct {
	config     Config
	httpClient *http.Client
	token      string
	tokenExp   time.Time
}

// NewGraphMailer crea un nuevo remitente con la configuración dada
func NewGraphMailer(cfg Config) *GraphMailer {
	return &GraphMailer{
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// getAccessToken obtiene un token OAuth2 usando client_credentials
func (g *GraphMailer) getAccessToken(ctx context.Context) (string, error) {
	// Si el token actual es válido (con margen de 5 min), lo reutiliza
	if g.token != "" && time.Now().Add(5*time.Minute).Before(g.tokenExp) {
		return g.token, nil
	}

	url := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", g.config.TenantID)
	data := fmt.Sprintf("client_id=%s&client_secret=%s&scope=https://graph.microsoft.com/.default&grant_type=client_credentials",
		g.config.ClientID, g.config.ClientSecret)

	fmt.Println(data)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error obteniendo token %s", resp.Status)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	g.token = tokenResp.AccessToken
	g.tokenExp = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return g.token, nil
}

// Send envía un correo usando Microsoft Graph API
// fromAddress: dirección del remitente (ej. lautaro.teta@rbaresidences.com)
// to: lista de destinatarios (strings)
// subject, body, isHTML: contenido del mensaje
func (g *GraphMailer) Send(ctx context.Context, fromAddress string, to []string, subject, body string, isHTML bool) error {
	token, err := g.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("error autenticando: %w", err)
	}

	// Construir el mensaje en el formato Graph
	msg := Message{
		Subject: subject,
		Body: MessageBody{
			ContentType: "HTML",
			Content:     body,
		},
		ToRecipients:  mapStringsToRecipients(to),
	}

	payload, err := json.Marshal(map[string]interface{}{
		"message":          msg,
		"saveToSentItems":  true,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/sendMail", fromAddress)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		// Leer cuerpo del error para depurar
		var errBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errBody)
		return fmt.Errorf("graph api error: status %s, body %v", resp.Status, errBody)
	}

	return nil
}

// Helper para convertir []string en []Recipient
func mapStringsToRecipients(list []string) []Recipient {
	if len(list) == 0 {
		return nil
	}
	recips := make([]Recipient, len(list))
	for i, addr := range list {
		recips[i] = Recipient{
			EmailAddress: EmailAddress{Address: addr},
		}
	}
	return recips
}
