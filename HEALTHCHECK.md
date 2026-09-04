# Healthcheck de dependencias

Sistema para validar la salud de las APIs externas de LeadsExtractor, y del ecosistema Rebora completo (LeadsExtractor + Portalia + Cotizador).

---

## 1. Arquitectura

`pkg/dependency` define el contrato genérico:

```go
type Dependency interface {
    Name() string
    HealthCheck(ctx context.Context) error
}
```

Cada cliente externo lo implementa (`whatsapp.Whatsapp`, `infobip.InfobipApi`, `pipedrive.Pipedrive`, `email.GraphMailer`, `zenrows.ZenRows`, y `store.DBDependency` para MySQL). `dependency.CheckAll(ctx, deps, timeout)` los corre todos en paralelo y devuelve `[]Result`:

```go
type Result struct {
    Name      string
    Status    Status // "ok" | "warning" | "error"
    Error     string
    Warning   string
    LatencyMS int64
}
```

**`Status: "warning"`**: algunas dependencias (ZenRows, Infobip) implementan además `UsageReporter.UsageWarning(ctx)`, que avisa cuando se está por agotar el plan/saldo **antes** de que falle (ver sección 4). Un warning no cuenta como falla para `AllOK`.

`App.Dependencies()` (`bootstrap/app.go`) arma la lista con las 6 dependencias propias de LeadsExtractor. Esto **no** es lo mismo que "todo el ecosistema Rebora" — ver sección 3.

## 2. Cada dependencia, cómo se valida

Ninguno de estos chequeos tiene efectos secundarios (no manda mensajes reales, no gasta créditos de scraping, no crea recursos):

| Dependencia | Cómo se valida | Archivo |
|---|---|---|
| MySQL | `db.PingContext` | `store/health.go` |
| WhatsApp/Meta | `GET /{numberId}?fields=verified_name` con el access token | `pkg/whatsapp/api.go` |
| Infobip | `GET /account/1/balance` | `pkg/infobip/api.go` |
| Pipedrive | `GET users/me` con el ApiToken | `pkg/pipedrive/api.go` |
| Microsoft Graph Mail | Obtiene un access token OAuth (sin mandar mail) | `pkg/email/graphmail.go` |
| ZenRows | `GET /v1/subscriptions/self/details` (no consume crédito, documentado por ZenRows) | `pkg/zenrows/api.go` |

**Nota Infobip**: en este momento devuelve 401 "Invalid login details" en cualquier GET, incluso en endpoints donde la key sí tiene permiso de POST (usado en producción) — todo apunta a **IP allowlisting** en la key de Infobip, no a un bug del código. Pendiente de confirmar/arreglar del lado de Infobip.

## 3. Ecosistema Rebora completo (`App.CheckRebora`)

`bootstrap/app.go`, método `CheckRebora(ctx)`: combina `Dependencies()` propio + el `/health` de Portalia (desglosado campo por campo) + un chequeo simple de 200 a Cotizador.

```go
func (a *App) CheckRebora(ctx context.Context) []dependency.Result {
    results := dependency.CheckAll(ctx, a.Dependencies(), 10*time.Second)
    results = append(results, dependency.FetchRemoteHealth(ctx, "portalia", os.Getenv("PORTALIA_URL")+"/api/v1/health", 30*time.Second)...)
    results = append(results, dependency.CheckURL(ctx, "cotizador", os.Getenv("COTIZADOR_URL"), 10*time.Second))
    return results
}
```

- **`dependency.FetchRemoteHealth`** (`pkg/dependency/remote.go`): le pega al `/health` de otro servicio que exponga el mismo contrato (`{success, data: [{name, status, error?, warning?, latency_ms}]}`) y "aplana" cada entrada con un prefijo (ej. `portalia.scrapers.wordpress`), en vez de colapsar todo en un único "portalia: error" genérico. Si Portalia mismo está inalcanzable, cae a un único resultado de error.
- **`dependency.CheckURL`**: para servicios que no tienen `/health` estructurado (Cotizador) — solo confirma 200.

**Env vars necesarias**: `PORTALIA_URL` (sin `/api/v1`, se agrega solo) y `COTIZADOR_URL`. Si no están seteadas, esos chequeos simplemente no corren (no se reportan como caídos).

### Dónde se usa `CheckRebora`

- `GET /health/rebora` — todo el ecosistema.
- WhatsApp, mensaje exacto `"healthcheck"` desde un número en `REPORT_NUMBERS` (ver sección 5).
- `go run ./cmd/healthcheck` — mismo resultado por consola, exit code 1 si algo falló.

`GET /health` (sin `/rebora`) sigue siendo **solo LeadsExtractor** — útil si algo (un monitor externo) ya apunta ahí y no querés que un problema en Portalia/Cotizador lo tire abajo también.

## 4. Warnings de consumo (antes de que falle)

No es posible saber por API si un método de pago está por vencer (investigado a fondo para Meta/WhatsApp: no existe ningún endpoint de Graph API para eso, ni con permisos de admin — solo lo avisa Meta por mail). Lo que sí se puede anticipar es el consumo de plan/saldo:

- **ZenRows**: si `usage_percent >= 85%`, warning con la fecha en que vence el período.
- **Infobip**: si el balance (`/account/1/balance`) cae debajo de 10, warning (bloqueado hoy por el mismo problema de permisos de la sección 2).

## 5. Trigger de WhatsApp

`flow/actions.go`, acción `health.check`, registrada como cualquier otra acción del sistema de flows (`flow.DefineActions`). Se dispara con una regla en `actions.json` (o vía la API `PUT /flows/{uuid}` en producción) con `"condition": {"message": "healthcheck"}`.

**Restricción de acceso**: solo responde si el número que escribe está en `REPORT_NUMBERS` (env var, separada por `;`, mismo formato que usa `cmd/reporter`). Si no está autorizado, la acción no hace nada — no expone el estado de las APIs a un lead cualquiera. Como esa regla se evalúa antes que el catch-all del flow, alguien no autorizado que por casualidad escriba "healthcheck" no recibe ni el saludo automático ni el healthcheck (silencio) — caso límite aceptado, no una palabra que un lead real vaya a escribir.

## 6. Endpoints HTTP

| Ruta | Qué chequea |
|---|---|
| `GET /health` | Solo LeadsExtractor |
| `GET /health/rebora` | LeadsExtractor + Portalia (por portal) + Cotizador |

Ambos devuelven 200 si todo OK, 503 si algo tiene `status: "error"` (los warnings no bajan el status HTTP).

## 7. CLI

```
cd src && go run ./cmd/healthcheck
```

Mismo resultado que `/health/rebora`, en texto plano, con exit code 1 si algo falló — pensado para correr en un cron/monitoreo externo.

## 8. Reporte diario (relacionado, no es healthcheck)

El reporte diario de WhatsApp (`service/report.go`) dejó de depender de editar un template de Meta cada vez que se agrega una fuente de lead. Documentado en detalle en `WHATSAPP_TEMPLATES.md`.

## 9. Monitoreo externo (pendiente / a definir)

Evaluado UptimeRobot (polling activo a una URL, plan gratis) vs. Healthchecks.io (dead man's switch, necesita que algo de este lado le haga ping — no encaja sin un cron intermedio). Se optó por construir el chequeo detallado acá mismo (`CheckRebora`) en vez de depender de que un monitor externo entienda la estructura de cada servicio.

**Sin resolver todavía**: alerta *proactiva* (que avise sola, sin que alguien pregunte por WhatsApp o pegue a `/health/rebora`). La idea con la que quedó, para cuando se retome: un endpoint relay en LeadsExtractor que reciba un webhook de UptimeRobot y lo reenvíe por WhatsApp usando un template de un solo parámetro (mismo patrón que `reporte_dinamico`, ver `WHATSAPP_TEMPLATES.md`), ya que un mensaje de WhatsApp iniciado por la empresa necesita salir por template, no texto libre.
