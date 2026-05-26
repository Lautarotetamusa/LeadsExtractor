# Refactor: Módulo Messenger

## Objetivo

Extraer toda la lógica de mensajería, flows y respuestas automáticas del servicio principal (`api/`) a un módulo independiente (`messenger`). Este módulo es el único responsable de decidir qué mensajes enviar y cuándo, reemplazando el sistema actual basado en `api/flow/` y la tabla `Action`.

---

## Problema con el sistema actual

- Los mensajes programados viven en memoria (`atomicgo.dev/schedule`). Si el proceso se reinicia, se pierden.
- El máximo práctico para programar un mensaje es ~24h (limitación del scheduler en memoria).
- La lógica de flows, acciones y mensajería está acoplada al servicio HTTP principal.
- La tabla `Action` registra qué acción se ejecutó pero no guarda el contenido del mensaje enviado, dificultando el historial.

---

## Solución

Un binario separado (`cmd/messenger`) con su propia tabla `Messages` en la DB. Los mensajes se persisten antes de enviarse, lo que permite programarlos a cualquier tiempo futuro y sobrevivir reinicios.

---

## Arquitectura del módulo

```
api/
  cmd/messenger/main.go
  messenger/
    flow/
      types.go        ← Flow, Rule, Action, Condition
      manager.go      ← carga flows desde JSON
    service/
      service.go      ← HandleIncoming, runFlow, scheduleAction
      scheduler.go    ← poller DB → envío
    store/
      messages.go     ← modelo Message + CRUD
    handler/
      handler.go      ← rutas HTTP
db/
  14_messages.sql     ← migración
```

---

## Tabla Messages

```sql
CREATE TABLE Messages (
    id           BIGINT      AUTO_INCREMENT PRIMARY KEY,
    phone        CHAR(16)    NOT NULL,
    type         VARCHAR(32) NOT NULL DEFAULT 'text',
    content      TEXT        NOT NULL,
    outgoing     BOOLEAN     NOT NULL DEFAULT FALSE,
    scheduled_at DATETIME    NOT NULL,
    sended_at    DATETIME    DEFAULT NULL,   -- NULL = pendiente de envío
    on_response  CHAR(37)    DEFAULT NULL,   -- UUID del flow a correr si el lead responde

    INDEX idx_phone   (phone),
    INDEX idx_pending (outgoing, scheduled_at, sended_at)
);
```

Reemplaza a `Action` (que registraba qué acción corrió) y complementa a `Message` (que guardaba mensajes entrantes). Unifica entrantes y salientes en una sola tabla.

---

## Flujo de un mensaje entrante

```
POST /incoming  ó  POST /webhooks
        │
        ▼
HandleIncoming(payload)
        │
        ├── INSERT Messages (outgoing=false, sended_at=now)
        │
        ├── busca último outgoing con on_response para ese phone
        │       ├── existe  → corre ese flow
        │       └── no existe → corre el flow main
        │
        └── runFlow → por cada action del rule matching:
                INSERT Messages (outgoing=true, scheduled_at=now+interval, sended_at=NULL)
```

```
Scheduler (cada 5s)
        │
        ├── SELECT WHERE outgoing=TRUE AND scheduled_at<=NOW() AND sended_at IS NULL
        │
        └── por cada mensaje pendiente:
                ├── envía por WhatsApp (text o template)
                └── UPDATE sended_at=NOW()
```

---

## Rutas expuestas

| Método | Ruta                 | Descripción                                        |
|--------|----------------------|----------------------------------------------------|
| POST   | `/incoming`          | Recibe un lead de portal, scraper o sistema externo |
| POST   | `/webhooks`          | Webhook de WhatsApp (mensajes entrantes)           |
| GET    | `/webhooks`          | Verificación del webhook de WhatsApp               |
| POST   | `/messages/schedule` | Programa un mensaje para una fecha futura          |
| GET    | `/messages/pending`  | Lista mensajes pendientes de envío                 |

### Payload `POST /incoming`

```json
{
  "phone":   "5491112345678",
  "name":    "Juan Pérez",
  "content": "Hola, me interesa la propiedad",
  "source":  "inmuebles24",
  "is_new":  true
}
```

### Payload `POST /messages/schedule`

```json
{
  "phone":        "5491112345678",
  "type":         "wpp.message",
  "content":      "Hola {{.Name}}, te escribimos de nuevo",
  "scheduled_at": "2026-05-01T09:00:00Z",
  "on_response":  "uuid-del-flow-opcional"
}
```

---

## Formato de flows (JSON)

El módulo reutiliza el mismo `actions.json` existente. La única diferencia es que las condiciones ahora matchean campos del `IncomingPayload` en vez del modelo `Communication`.

```json
{
  "Main": "<uuid>",
  "Flows": {
    "<uuid>": {
      "Name": "Bienvenida",
      "Rules": [
        {
          "condition": {
            "is_new": true,
            "source": "inmuebles24"
          },
          "actions": [
            {
              "action": "wpp.message",
              "interval": "0s",
              "params": { "text": "¡Hola {{.Name}}! Recibimos tu consulta." }
            },
            {
              "action": "wpp.message",
              "interval": "24h",
              "params": { "text": "¿Querés que te pasemos más info?" }
            }
          ],
          "on_response": "<uuid-del-flow-de-respuesta>"
        }
      ]
    }
  }
}
```

### Tipos de acción soportados

| Acción          | Comportamiento                                         |
|-----------------|--------------------------------------------------------|
| `wpp.message`   | Texto renderizado con Go templates (`{{.Name}}`, etc.) |
| `wpp.template`  | Template de WhatsApp Business (params como JSON)       |

---

## Variables de entorno

```
MESSENGER_PORT=8081
MESSENGER_FLOWS_PATH=../actions.json
WHATSAPP_ACCESS_TOKEN=...
WHATSAPP_NUMBER_ID=...
WHATSAPP_VERIFY_TOKEN=...
```

Las variables de DB (`HOST`, `DB_USER`, `DB_PASS`, `DB_PORT`, `DB_NAME`) son las mismas que el servicio principal — comparten la misma base de datos.

---

## Migración

```bash
mysql -u <user> -p LeadsExtractor < db/14_messages.sql
```

---

## Qué queda por hacer

- [ ] Implementar `GET /messages/pending` (actualmente devuelve 501)
- [ ] Soporte para acciones no-mensajería dentro de flows (`pipedrive.save`, `infobip.save`). Opciones: webhook de callback al servicio principal, o registrar handlers en el módulo.
- [ ] Migrar el servicio principal para que llame a `POST /incoming` en vez de correr flows internamente
- [ ] Eliminar `api/flow/` y la tabla `Action` una vez validado el nuevo módulo en producción
- [ ] Manejo de reintentos para mensajes que fallan al enviarse
