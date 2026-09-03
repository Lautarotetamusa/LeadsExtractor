package service

import (
	"context"
	"fmt"
	"leadsextractor/pkg/email"
	"leadsextractor/pkg/whatsapp"
	"leadsextractor/store"
	"strings"
	"time"
)

type ReportService struct {
	commStore store.CommunicationStorer
	wa        *whatsapp.Whatsapp
	mailer    email.Sender
}

type SourceReportData struct {
	Prospected int
	NewLeads   int
	Existing   int
	Total      int
}

// reportedSources son las fuentes que se incluyen en el reporte diario (por
// WhatsApp y por email). Agregar una fuente nueva acá no requiere tocar
// ningún template de Meta: el mensaje de WhatsApp se arma dinámicamente en
// un solo parámetro de texto (ver buildReportText).
var reportedSources = []string{"inmuebles24", "casasyterrenos", "lamudi", "propiedades", "whatsapp", "easybroker", "wordpress"}

func NewReportService(store store.CommunicationStorer, wa *whatsapp.Whatsapp, mailer email.Sender) *ReportService {
	return &ReportService{commStore: store, wa: wa, mailer: mailer}
}

// Genera estadísticas diarias
func (rs *ReportService) GenerateDailyStats(date time.Time) (map[string]SourceReportData, error) {
    // Ajustar a UTC para evitar problemas de zona horaria
    utcLoc := time.UTC
    start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location()).In(utcLoc)
    end := time.Now().AddDate(0, 0, 1).In(utcLoc) // tomorrow

    stats, err := rs.commStore.GetCommunicationStats(start, end)
    if err != nil {
        return nil, err
    }

    // Agrupar por fuente
    sourceMap := make(map[string]SourceReportData)

    for _, row := range stats {
        data := sourceMap[row.Source]
        if row.NewLead {
            data.NewLeads += row.Count
        } else {
            data.Existing += row.Count
        }
		if row.UtmSource.Valid && row.UtmSource.String == "prospectador" {
			data.Prospected += row.Count
		}

        data.Total += row.Count
        sourceMap[row.Source] = data
    }

	return sourceMap, nil
}

func (fs *ReportService) SendDailyReport(numbers []string) error {
	return fs.SendReport(numbers, 1) // 1 day before
}

// Envía reporte a múltiples números
func (rs *ReportService) SendReport(numbers []string, daysBefore int) error {
    date := time.Now().AddDate(0, 0, -daysBefore)
    report, err := rs.GenerateDailyStats(date)

    if err != nil {
        return err
    }

	text := buildReportText(date, reportedSources, report)

	// El template "reporte_dinamico" tiene un único parámetro {{1}} en el
	// body. Agregar/sacar una fuente del reporte es 100% código Go (ver
	// reportedSources/buildReportText); no requiere volver a aprobar nada
	// en Meta, porque la cantidad de parámetros del template nunca cambia.
	t := whatsapp.TemplatePayload{
		Name: "reporte_dinamico",
		Language: whatsapp.Language{
			Code: "es_MX",
		},
		Components: []whatsapp.Components{
			{
				Type: "body",
				Parameters: []whatsapp.Parameter{
					{Type: "text", Text: text},
				},
			},
		},
	}

	for _, number := range numbers {
		if _, err := rs.wa.SendTemplate(number, t); err != nil {
			return fmt.Errorf("error enviando a %s: %w", number, err)
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

// buildReportText arma el reporte como un único renglón de texto: los
// parámetros de un template de WhatsApp no admiten saltos de línea ni tabs
// (la Cloud API los rechaza), así que las fuentes se separan con " | " en
// lugar de un salto de línea por fuente.
func buildReportText(date time.Time, sources []string, report map[string]SourceReportData) string {
	rows := make([]string, 0, len(sources))
	total := 0
	for _, source := range sources {
		d := report[source]
		rows = append(rows, fmt.Sprintf("%s: %d nuevos, %d existentes, %d total", source, d.NewLeads, d.Existing, d.Total))
		total += d.Total
	}

	return fmt.Sprintf(
		"%s | %s | Total: %d | Prospectados (i24): %d",
		date.Format("2006-01-02"),
		strings.Join(rows, " | "),
		total,
		report["inmuebles24"].Prospected,
	)
}

func (rs *ReportService) SendDailyReportEmail(recipients []string) error {
	return rs.SendReportEmail(recipients, 1)
}

func (rs *ReportService) SendReportEmail(recipients []string, daysBefore int) error {
	if rs.mailer == nil {
		return fmt.Errorf("mailer no configurado")
	}

	date := time.Now().AddDate(0, 0, -daysBefore)
	report, err := rs.GenerateDailyStats(date)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("Reporte diario - %s", date.Format("2006-01-02"))
	body := buildReportHTML(date, reportedSources, report)

	return rs.mailer.Send(context.Background(), recipients, subject, body, true)
}

func buildReportHTML(date time.Time, sources []string, report map[string]SourceReportData) string {
	var sb strings.Builder

	sb.WriteString(`<html><body>`)
	sb.WriteString(fmt.Sprintf(`<h2>Reporte del %s</h2>`, date.Format("2006-01-02")))
	sb.WriteString(`<table border="1" cellpadding="6" cellspacing="0" style="border-collapse:collapse;font-family:sans-serif">`)
	sb.WriteString(`<thead><tr style="background:#f0f0f0">`)
	sb.WriteString(`<th>Fuente</th><th>Nuevos</th><th>Existentes</th><th>Total</th><th>Prospectados</th></tr></thead><tbody>`)

	grandTotal := 0
	for _, source := range sources {
		d := report[source]
		prospected := "-"
		if source == "inmuebles24" {
			prospected = fmt.Sprintf("%d", d.Prospected)
		}
		sb.WriteString(fmt.Sprintf(
			`<tr><td>%s</td><td>%d</td><td>%d</td><td>%d</td><td>%s</td></tr>`,
			source, d.NewLeads, d.Existing, d.Total, prospected,
		))
		grandTotal += d.Total
	}

	sb.WriteString(fmt.Sprintf(
		`<tr style="font-weight:bold;background:#f0f0f0"><td>TOTAL</td><td></td><td></td><td>%d</td><td>%d</td></tr>`,
		grandTotal, report["inmuebles24"].Prospected,
	))
	sb.WriteString(`</tbody></table>`)

	sb.WriteString(`</body></html>`)
	return sb.String()
}
