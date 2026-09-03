// Comando standalone para validar todo el ecosistema Rebora: las APIs
// externas propias (ZenRows, WhatsApp, Infobip, Pipedrive, MySQL, Microsoft
// Graph Mail), más Portalia (desglosado por portal) y Cotizador. Termina
// con exit code 1 si algo falla, para poder usarse en un cron/monitoreo.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"leadsextractor/bootstrap"
	"leadsextractor/pkg/dependency"
)

func main() {
	ctx := context.Background()

	app, err := bootstrap.NewApp(ctx)
	if err != nil {
		log.Fatal(err)
	}

	results := app.CheckRebora(ctx)

	failed := false
	for _, r := range results {
		switch r.Status {
		case dependency.StatusOK:
			fmt.Printf("[OK]      %-25s %dms\n", r.Name, r.LatencyMS)
		case dependency.StatusWarning:
			fmt.Printf("[WARNING] %-25s %dms: %s\n", r.Name, r.LatencyMS, r.Warning)
		default:
			failed = true
			fmt.Printf("[ERROR]   %-25s %dms: %s\n", r.Name, r.LatencyMS, r.Error)
		}
	}

	if failed {
		os.Exit(1)
	}
}
