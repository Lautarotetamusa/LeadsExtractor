package handlers

import (
    "fmt"
    "time"
    "leadsextractor/pkg/whatsapp"
    "leadsextractor/store"
)

type ReportService struct {
    commStore store.CommunicationStorer
    wa    *whatsapp.Whatsapp
}

type SourceReportData struct {
	NewLeads     int
	Existing     int
	Total        int
}

func NewReportService(store store.CommunicationStorer, wa *whatsapp.Whatsapp) *ReportService {
    return &ReportService{commStore: store, wa: wa}
}

// Genera estadísticas diarias
func (rs *ReportService) GenerateDailyStats(date time.Time) (map[string]SourceReportData, error) {
    // Ajustar a UTC para evitar problemas de zona horaria
    utcLoc := time.UTC
    start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location()).In(utcLoc)
    end := time.Now().AddDate(0, 0, 1).In(utcLoc) // tomorrow

    stats, err := rs.commStore.GetCommunicationStats(start, end)
	fmt.Printf("%#v\n", stats)
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
        data.Total += row.Count
        sourceMap[row.Source] = data
    }

    // Construir reporte
    // report := "📊 *Reporte Diario de Comunicaciones*\n"
    // report += fmt.Sprintf("Fecha desde: %s\n\n", date.Format("2006-01-02"))
    //
    // total := 0
    // for source, data := range sourceMap {
    //     report += fmt.Sprintf("*%s:*\n", source)
    //     report += fmt.Sprintf("  - Nuevos: %d\n", data.NewLeads)
    //     report += fmt.Sprintf("  - Duplicados: %d\n", data.Existing)
    //     report += fmt.Sprintf("  - Total: %d\n\n", data.Total)
    //     total += data.Total
    // }
    // report += "-----------------------------\n"
    // report += fmt.Sprintf("TOTAL: %d\n\n", total)
    //
    // return report, nil
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

	// Same order than the template
	sendedSources := []string{"inmuebles24", "casasyterrenos", "lamudi", "propiedades", "whatsapp"}
	p := []whatsapp.Parameter{whatsapp.Parameter{
		Type: "text",
		Text: date.Format("2006-01-02"),
	}}

	total := 0
	for _, source := range sendedSources {
		p = append(p, whatsapp.Parameter{
			Type: "text",
			Text: fmt.Sprintf("%d", report[source].NewLeads),
		})
		p = append(p, whatsapp.Parameter{
			Type: "text",
			Text: fmt.Sprintf("%d", report[source].Existing),
		})
		p = append(p, whatsapp.Parameter{
			Type: "text",
			Text: fmt.Sprintf("%d", report[source].Total),
		})
		total += report[source].Total
	}
	p = append(p, whatsapp.Parameter{
		Type: "text",
		Text: fmt.Sprintf("%d", total),
	})
	fmt.Println(len(p))

	t := whatsapp.TemplatePayload {
		Name: "reporte",
		Language: whatsapp.Language{
			Code: "es_MX",
		},
		Components: []whatsapp.Components{
			{
				Type: "body",
				Parameters: p,
			},
		},
	}

    for _, number := range numbers {
        if _, err := rs.wa.SendTemplate(number, t); err != nil {
            return fmt.Errorf("error enviando a %s: %w", number, err)
        }
        time.Sleep(1 * time.Second) // Espacio entre envíos
    }
    
    return nil
}
