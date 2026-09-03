// Package zenrows es un cliente mínimo para la API de ZenRows
// (https://www.zenrows.com/), usado hoy solo para validar que el API key
// configurado es válido.
package zenrows

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// healthUrl devuelve el detalle de la suscripción/uso de la cuenta. ZenRows
// documenta que esta llamada no consume créditos de scraping ni cuenta
// contra el límite de concurrencia, por lo que es segura para un
// health-check.
const healthUrl = "https://api.zenrows.com/v1/subscriptions/self/details"

// usageWarnPercent es el umbral de uso del plan a partir del cual se avisa
// que se está por agotar antes de que el servicio empiece a fallar.
const usageWarnPercent = 85.0

type ZenRows struct {
	apiKey string
	client *http.Client
	logger *slog.Logger
}

type subscriptionDetails struct {
	UsagePercent float64 `json:"usage_percent"`
	PeriodEndsAt string  `json:"period_ends_at"`
}

func NewZenRows(apiKey string, l *slog.Logger) *ZenRows {
	return &ZenRows{
		apiKey: apiKey,
		client: &http.Client{Timeout: 15 * time.Second},
		logger: l.With("module", "zenrows"),
	}
}

func (z *ZenRows) getSubscription(ctx context.Context) (*subscriptionDetails, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-API-Key", z.apiKey)

	res, err := z.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("no se pudo realizar la peticion: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("respuesta inesperada de zenrows: %d", res.StatusCode)
	}

	var details subscriptionDetails
	if err := json.NewDecoder(res.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("no se pudo parsear la respuesta de zenrows: %w", err)
	}
	return &details, nil
}

// UsageWarning avisa cuando el consumo del plan supera usageWarnPercent, para
// enterarse antes de que ZenRows empiece a rechazar requests por falta de
// crédito/renovación de plan.
func (z *ZenRows) UsageWarning(ctx context.Context) (string, error) {
	details, err := z.getSubscription(ctx)
	if err != nil {
		return "", err
	}

	if details.UsagePercent >= usageWarnPercent {
		return fmt.Sprintf("uso del plan al %.1f%%, vence %s", details.UsagePercent, details.PeriodEndsAt), nil
	}
	return "", nil
}

func (z *ZenRows) Name() string {
	return "zenrows"
}

func (z *ZenRows) HealthCheck(ctx context.Context) error {
	_, err := z.getSubscription(ctx)
	return err
}
