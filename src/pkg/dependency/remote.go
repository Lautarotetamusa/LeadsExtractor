package dependency

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type remoteResult struct {
	Name      string `json:"name"`
	Status    Status `json:"status"`
	Error     string `json:"error,omitempty"`
	Warning   string `json:"warning,omitempty"`
	LatencyMS int64  `json:"latency_ms"`
}

type remoteHealthResponse struct {
	Success bool           `json:"success"`
	Data    []remoteResult `json:"data"`
}

// FetchRemoteHealth le pega al /health de otro servicio de Rebora (que
// expone el mismo contrato {success, data: [{name, status, error, warning,
// latency_ms}]}) y "aplana" cada entrada de su respuesta en el resultado
// combinado, prefijada con prefix (ej. "portalia.mysql", "portalia.scrapers.
// wordpress"), para no perder el detalle de qué falló puntualmente en el
// otro servicio.
//
// Si no se puede ni siquiera contactar al servicio (down, timeout, DNS,
// respuesta no-JSON), devuelve un único Result de error con name=prefix.
func FetchRemoteHealth(ctx context.Context, prefix string, url string, timeout time.Duration) []Result {
	start := time.Now()

	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(cctx, http.MethodGet, url, nil)
	if err != nil {
		return []Result{unreachable(prefix, start, err)}
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return []Result{unreachable(prefix, start, err)}
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return []Result{unreachable(prefix, start, err)}
	}

	var parsed remoteHealthResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return []Result{unreachable(prefix, start, fmt.Errorf("respuesta invalida (%d): %s", res.StatusCode, body))}
	}

	if len(parsed.Data) == 0 {
		status := StatusOK
		if res.StatusCode != http.StatusOK {
			status = StatusError
		}
		return []Result{{
			Name:      prefix,
			Status:    status,
			LatencyMS: time.Since(start).Milliseconds(),
		}}
	}

	results := make([]Result, len(parsed.Data))
	for i, r := range parsed.Data {
		results[i] = Result{
			Name:      prefix + "." + r.Name,
			Status:    r.Status,
			Error:     r.Error,
			Warning:   r.Warning,
			LatencyMS: r.LatencyMS,
		}
	}
	return results
}

// CheckURL confirma que una URL responde 200, sin esperar ningún formato de
// respuesta particular (para servicios que no exponen su propio /health
// estructurado, ej. Cotizador).
func CheckURL(ctx context.Context, name string, url string, timeout time.Duration) Result {
	start := time.Now()

	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(cctx, http.MethodGet, url, nil)
	if err != nil {
		return unreachable(name, start, err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return unreachable(name, start, err)
	}
	defer res.Body.Close()

	latency := time.Since(start).Milliseconds()
	if res.StatusCode != http.StatusOK {
		return Result{
			Name:      name,
			Status:    StatusError,
			Error:     fmt.Sprintf("respuesta inesperada: %d", res.StatusCode),
			LatencyMS: latency,
		}
	}

	return Result{Name: name, Status: StatusOK, LatencyMS: latency}
}

func unreachable(name string, start time.Time, err error) Result {
	return Result{
		Name:      name,
		Status:    StatusError,
		Error:     err.Error(),
		LatencyMS: time.Since(start).Milliseconds(),
	}
}
