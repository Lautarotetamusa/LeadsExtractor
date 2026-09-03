// Comando standalone para validar todas las APIs externas de las que
// depende la aplicación (ZenRows, WhatsApp, Infobip, Pipedrive, MySQL,
// Microsoft Graph Mail). Termina con exit code 1 si alguna falla, para
// poder usarse en un cron/monitoreo.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"leadsextractor/bootstrap"
	"leadsextractor/pkg/dependency"
)

func main() {
	ctx := context.Background()

	app, err := bootstrap.NewApp(ctx)
	if err != nil {
		log.Fatal(err)
	}

	results := dependency.CheckAll(ctx, app.Dependencies(), 15*time.Second)

	failed := false
	for _, r := range results {
		switch r.Status {
		case dependency.StatusOK:
			fmt.Printf("[OK]      %-12s %s\n", r.Name, r.Latency)
		case dependency.StatusWarning:
			fmt.Printf("[WARNING] %-12s %s: %s\n", r.Name, r.Latency, r.Warning)
		default:
			failed = true
			fmt.Printf("[ERROR]   %-12s %s: %s\n", r.Name, r.Latency, r.Error)
		}
	}

	if failed {
		os.Exit(1)
	}
}
