package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const apiURL = "https://api.postmarkapp.com/email"

type Client struct {
	serverToken string
	from        string
	httpClient  *http.Client
}

func NewClient(serverToken, from string) *Client {
	return &Client{
		serverToken: serverToken,
		from:        from,
		httpClient:  &http.Client{},
	}
}

type sendRequest struct {
	From          string `json:"From"`
	To            string `json:"To"`
	Subject       string `json:"Subject"`
	HtmlBody      string `json:"HtmlBody"`
	MessageStream string `json:"MessageStream"`
}

type sendResponse struct {
	ErrorCode int    `json:"ErrorCode"`
	Message   string `json:"Message"`
}

func (c *Client) Send(to []string, subject, body string) error {
	payload := sendRequest{
		From:          c.from,
		To:            strings.Join(to, ","),
		Subject:       subject,
		HtmlBody:      body,
		MessageStream: "outbound",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("postmark: error serializando payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("postmark: error creando request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", c.serverToken)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("postmark: error enviando request: %w", err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("postmark: error leyendo respuesta: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		var postmarkErr sendResponse
		if err := json.Unmarshal(resBody, &postmarkErr); err == nil && postmarkErr.Message != "" {
			return fmt.Errorf("postmark error %d: %s", postmarkErr.ErrorCode, postmarkErr.Message)
		}
		return fmt.Errorf("postmark: status %d: %s", res.StatusCode, string(resBody))
	}

	return nil
}
