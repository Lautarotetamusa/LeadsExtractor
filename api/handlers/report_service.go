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

func NewReportService(store store.CommunicationStorer, wa *whatsapp.Whatsapp) *ReportService {
    return &ReportService{commStore: store, wa: wa}
}

// Genera estadísticas diarias
func (rs *ReportService) GenerateDailyStats(date time.Time) (string, error) {
    // Ajustar a UTC para evitar problemas de zona horaria
    utcLoc := time.UTC
    start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location()).In(utcLoc)
    end := time.Now().AddDate(0, 0, 1).In(utcLoc) // tomorrow

    stats, err := rs.commStore.GetCommunicationStats(start, end)
    if err != nil {
        return "", err
    }

    // Agrupar por fuente
    sourceMap := make(map[string]struct {
        NewLeads     int
        Existing     int
        Total        int
    })

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
    report := "📊 *Reporte Diario de Comunicaciones*\n"
    report += fmt.Sprintf("Fecha desde: %s\n\n", date.Format("2006-01-02"))
    
    total := 0
    for source, data := range sourceMap {
        report += fmt.Sprintf("*%s:*\n", source)
        report += fmt.Sprintf("  - Nuevos: %d\n", data.NewLeads)
        report += fmt.Sprintf("  - Duplicados: %d\n", data.Existing)
        report += fmt.Sprintf("  - Total: %d\n\n", data.Total)
        total += data.Total
    }
    report += "-----------------------------\n"
    report += fmt.Sprintf("TOTAL: %d\n\n", total)

    return report, nil
}

// Envía reporte a múltiples números
func (rs *ReportService) SendDailyReport(numbers []string) error {
    reportDate := time.Now().AddDate(0, 0, -1) // Estadísticas del día anterior
    report, err := rs.GenerateDailyStats(reportDate)

    if err != nil {
        return err
    }

    for _, number := range numbers {
        if _, err := rs.wa.SendMessage(number, report); err != nil {
            return fmt.Errorf("error enviando a %s: %w", number, err)
        }
        time.Sleep(1 * time.Second) // Espacio entre envíos
    }
    
    return nil
}
