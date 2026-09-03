# Templates de WhatsApp requeridos

Mensajes que el sistema envía fuera de una sesión de 24hs (business-initiated) y que por eso necesitan un template aprobado por Meta. Se crean en **WhatsApp Manager → Message Templates** del Business Manager.

---

## `reporte_dinamico`

Usado por el reporte diario (`cmd/reporter`, `service/report.go`) para avisar por WhatsApp las métricas de leads del día anterior a los números en `REPORT_NUMBERS`.

Reemplaza al viejo template `reporte`, que tenía un parámetro fijo por cada fuente/métrica (`{{1}}`...`{{18}}`) y que había que editar en Meta cada vez que se agregaba o sacaba una fuente. Este template tiene **un solo parámetro** que recibe el reporte ya armado como texto — agregar una fuente nueva a futuro es un cambio de código en `service/report.go` (variable `reportedSources`), sin volver a tocar Meta.

**Configuración:**

| Campo | Valor |
|---|---|
| Nombre | `reporte_dinamico` |
| Categoría | Utility |
| Idioma | `es_MX` |

**Body:**

```
{{1}}
```

**Ejemplo de contenido de muestra** (para completar el campo de ejemplo que pide Meta al armar el template):

```
2026-09-03 | inmuebles24: 5 nuevos, 2 existentes, 7 total | casasyterrenos: 1 nuevos, 0 existentes, 1 total | lamudi: 2 nuevos, 1 existentes, 3 total | propiedades: 0 nuevos, 0 existentes, 0 total | whatsapp: 8 nuevos, 4 existentes, 12 total | easybroker: 2 nuevos, 0 existentes, 2 total | wordpress: 1 nuevos, 1 existentes, 2 total | Total: 27 | Prospectados (i24): 3
```

**Por qué es una sola línea:** la Cloud API de Meta rechaza saltos de línea y tabs dentro del *valor* de un parámetro (error 400: *"Param text cannot have new-line/tab characters or more than 4 consecutive spaces"*), aunque sí se puedan usar `\n` en el texto estático del template. Por eso `buildReportText` (en `service/report.go`) separa cada fuente con ` | ` en vez de un salto de línea.

**Una vez aprobado**, no hace falta volver a editarlo nunca — el código ya está preparado para usarlo tal cual (`whatsapp.TemplatePayload{Name: "reporte_dinamico", ...}`).

---

## `reporte` (deprecado)

El template viejo, con un parámetro por fuente/métrica. Ya no lo usa el código (reemplazado por `reporte_dinamico` arriba). Se puede desactivar/eliminar en Meta Business Manager una vez confirmado que `reporte_dinamico` funciona en producción.

---

## Otros templates existentes (sin cambios)

Estos ya estaban aprobados y el código los sigue usando sin modificaciones:

- `bienvenida_1`, `seguimiento_1`, `seguimiento_2`, `seguimiento_3` — secuencia de seguimiento de leads (`actions.json`, flow "Seguimientos").
- `info_asesor_2` — usado por `wpp.send_message_asesor` (`pkg/whatsapp/main.go`) para avisarle al asesor sobre un lead nuevo.
