// Package dependency define un contrato común para validar la salud de las
// APIs externas de las que depende la aplicación (ZenRows, Meta, Infobip,
// Pipedrive, MySQL, Microsoft Graph Mail), y una función para chequearlas
// todas juntas en paralelo.
package dependency

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Dependency lo implementa cada cliente externo (whatsapp.Whatsapp,
// infobip.InfobipApi, etc.) para poder ser validado de forma genérica.
type Dependency interface {
	// Name identifica la dependencia en el reporte (ej. "whatsapp", "infobip").
	Name() string
	// HealthCheck hace la petición mínima necesaria para confirmar que las
	// credenciales y la conectividad son válidas, sin efectos secundarios
	// (no envía mensajes, no gasta créditos de scraping, no crea recursos).
	HealthCheck(ctx context.Context) error
}

// UsageReporter lo implementan opcionalmente las dependencias que pueden
// avisar que se están por quedar sin cupo/saldo (ej. plan de ZenRows,
// balance de Infobip), para detectarlo antes de que el servicio empiece a
// fallar. warning queda vacío si no hay nada para avisar.
type UsageReporter interface {
	UsageWarning(ctx context.Context) (warning string, err error)
}

type Status string

const (
	StatusOK      Status = "ok"
	StatusWarning Status = "warning"
	StatusError   Status = "error"
)

type Result struct {
	Name    string        `json:"name"`
	Status  Status        `json:"status"`
	Error   string        `json:"error,omitempty"`
	Warning string        `json:"warning,omitempty"`
	Latency time.Duration `json:"latency_ms"`
}

// CheckAll corre el HealthCheck de cada dependencia en paralelo, acotado por
// timeout, y devuelve un resultado por dependencia (nunca devuelve error: las
// fallas individuales quedan reflejadas en cada Result).
func CheckAll(ctx context.Context, deps []Dependency, timeout time.Duration) []Result {
	results := make([]Result, len(deps))

	var wg sync.WaitGroup
	for i, d := range deps {
		wg.Add(1)
		go func(i int, d Dependency) {
			defer wg.Done()

			cctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			start := time.Now()
			err := d.HealthCheck(cctx)
			elapsed := time.Since(start)

			res := Result{
				Name:    d.Name(),
				Status:  StatusOK,
				Latency: elapsed,
			}
			if err != nil {
				res.Status = StatusError
				res.Error = err.Error()
				results[i] = res
				return
			}

			if reporter, ok := d.(UsageReporter); ok {
				warning, err := reporter.UsageWarning(cctx)
				if err != nil {
					res.Warning = fmt.Sprintf("no se pudo verificar el consumo: %s", err)
					res.Status = StatusWarning
				} else if warning != "" {
					res.Warning = warning
					res.Status = StatusWarning
				}
			}
			results[i] = res
		}(i, d)
	}
	wg.Wait()

	return results
}

// AllOK indica si ninguna dependencia falló duro (los warnings de consumo no
// cuentan como falla).
func AllOK(results []Result) bool {
	for _, r := range results {
		if r.Status == StatusError {
			return false
		}
	}
	return true
}
