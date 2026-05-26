# LeadsExtractor

Sistema de gestión y automatización de leads inmobiliarios. Captura leads de múltiples portales y vía WhatsApp, los asigna automáticamente a asesores mediante round-robin, y ejecuta flujos de comunicación automatizados a través de la WhatsApp Cloud API.

---

## Tabla de Contenidos

- [Descripción General](#descripción-general)
- [Arquitectura](#arquitectura)
- [Requerimientos](#requerimientos)
- [Variables de Entorno](#variables-de-entorno)
- [Modelos de Datos](#modelos-de-datos)
- [Esquema de Base de Datos](#esquema-de-base-de-datos)
- [Flujos (Flows)](#flujos-flows)
- [Integración WhatsApp](#integración-whatsapp)
- [API REST](#api-rest)
- [Procesamiento de Leads](#procesamiento-de-leads)
- [Integraciones](#integraciones)
- [Fuentes de Leads](#fuentes-de-leads)
- [Instalación](#instalación)
- [Testing de WhatsApp](#testing-de-whatsapp)

---

## Descripción General

LeadsExtractor automatiza el proceso de captación y seguimiento de leads inmobiliarios:

- **Captura** leads desde portales inmobiliarios (scraping Python) y vía WhatsApp (webhook)
- **Deduplica** leads por número de teléfono
- **Asigna** automáticamente leads a asesores mediante round-robin
- **Automatiza comunicaciones** vía WhatsApp usando un motor de flujos configurable (conditions + delays + templates)
- **Integra** con Pipedrive e Infobip como CRMs
- **Reporta** estadísticas diarias automáticas a los asesores

---

## Arquitectura

```
LeadsExtractor/
├── api/                        # Backend principal en Go
│   ├── main.go                 # Punto de entrada
│   ├── go.mod / go.sum
│   ├── actions.json            # Definición de flujos (flows)
│   ├── flow/                   # Motor de ejecución de flujos
│   ├── handlers/               # HTTP handlers
│   ├── models/                 # Structs de datos
│   ├── store/                  # Capa de base de datos (MySQL)
│   ├── middleware/             # Middleware HTTP
│   └── pkg/
│       ├── whatsapp/           # WhatsApp Cloud API (Meta)
│       ├── pipedrive/          # Pipedrive CRM (OAuth + API)
│       ├── infobip/            # Infobip mensajería
│       ├── roundrobin/         # Asignación round-robin
│       ├── jotform/            # Generación de cotizaciones PDF
│       └── logs/               # Logging estructurado
├── db/                         # Esquemas y migraciones SQL
├── messages/                   # Plantillas de mensajes
├── app/                        # Scrapers Python para portales
└── docker-compose.yml
```

**Stack:**
- **Backend**: Go con gorilla/mux
- **Base de datos**: MySQL (sqlx)
- **Mensajería**: WhatsApp Cloud API (Meta Graph API v17.0)
- **CRMs**: Pipedrive, Infobip
- **Scraping**: Python
- **Scheduler**: robfig/cron + atomicgo/schedule

---

## Requerimientos

- Go 1.21+
- MySQL 8+
- Cuenta de Meta Business con WhatsApp Cloud API habilitada
- (Opcional) Cuentas en Pipedrive e Infobip

### Dependencias Go principales

| Paquete | Uso |
|---------|-----|
| `gorilla/mux` | Routing HTTP |
| `jmoiron/sqlx` | SQL + named queries |
| `google/uuid` | Generación de UUIDs |
| `robfig/cron` | Reportes programados |
| `atomicgo.dev/schedule` | Delays en acciones de flujos |
| `nyaruka/phonenumbers` | Parseo y validación de teléfonos |
| `go-playground/validator` | Validación de structs |
| `samber/lo` | Utilidades funcionales |
| `lmittmann/tint` | Logging estructurado con color |

---

## Variables de Entorno

Crear `api/.env`:

```env
# Base de datos
DB_USER=
DB_PASS=
DB_PORT=3306
DB_NAME=
HOST=

# Servidor
API_PORT=8080
API_HOST=
APP_HOST=

# WhatsApp Cloud API (Meta)
WHATSAPP_ACCESS_TOKEN=
WHATSAPP_NUMBER_ID=
WHATSAPP_VERIFY_TOKEN=
WHATSAPP_IMAGE_ID=      # ID de imagen subida a Meta (para acción wpp.media)
WHATSAPP_VIDEO_ID=      # ID de video subido a Meta (para acción wpp.media)

# Pipedrive CRM
PIPEDRIVE_CLIENT_ID=
PIPEDRIVE_CLIENT_SECRET=
PIPEDRIVE_API_TOKEN=
PIPEDRIVE_REDIRECT_URI=

# Infobip
INFOBIP_APIKEY=
INFOBIP_APIURL=

# JotForm (generación de cotizaciones PDF)
JOTFORM_API_KEY=
JOTFORM_FORM_ID=

# Credenciales portales (scrapers Python)
CASASYTERRENOS_USERNAME=
CASASYTERRENOS_PASSWORD=
INMUEBLES24_USERNAME=
INMUEBLES24_PASSWORD=
LAMUDI_USERNAME=
LAMUDI_PASSWORD=
PROPIEDADES_USERNAME=
PROPIEDADES_PASSWORD=

# OneDrive (almacenamiento de imágenes de propiedades)
DRIVE_ID=

# Cron de reportes diarios (sintaxis cron estándar)
CRON="0 10-18/4 * * *"
```

---

## Modelos de Datos

### Communication

Unidad central del sistema. Representa cada consulta o contacto recibido.

```go
type Communication struct {
    Id         int
    Fuente     string       // "whatsapp", "ivr", "inmuebles24", "lamudi", etc.
    FechaLead  string       // Fecha y hora del contacto
    Nombre     string       // Nombre del contacto
    Link       string       // URL de origen (portal o enlace de WhatsApp)
    Telefono   PhoneNumber  // Teléfono del lead
    Email      NullString
    Utm        Utm          // Datos de tracking UTM
    Cotizacion string       // URL del PDF de cotización generado
    Asesor     Asesor       // Asesor asignado
    Propiedad  Propiedad    // Propiedad de interés
    Busquedas  Busquedas    // Criterios de búsqueda del lead
    IsNew      bool         // true = lead nuevo, false = duplicado
    Message    NullString   // Texto del mensaje recibido por WhatsApp
    Wamid      NullString   // ID del mensaje en WhatsApp (para tracking)
}
```

### Lead

Contacto único, deduplicado por número de teléfono.

```go
type Lead struct {
    Name       string
    Phone      PhoneNumber
    Email      NullString
    Asesor     Asesor
    Cotizacion string
}
```

### Asesor

Representante de ventas que recibe y atiende leads.

```go
type Asesor struct {
    Name   string
    Phone  PhoneNumber
    Email  string
    Active bool       // Participa en el round-robin cuando es true
}
```

### Propiedad

Inmueble de interés cuando el lead proviene de un portal.

```go
type Propiedad struct {
    ID          NullInt32
    Portal      string     // Portal de origen ("inmuebles24", "lamudi", etc.)
    PortalId    NullString // ID interno del portal
    Titulo      NullString
    Link        NullString
    Precio      NullString
    Ubicacion   NullString
    Tipo        NullString // "casa", "departamento", "terreno", etc.
    Bedrooms    string
    Bathrooms   string
    TotalArea   string
    CoveredArea string
}
```

### Busquedas

Criterios de búsqueda del lead registrados en el portal.

```go
type Busquedas struct {
    Zonas       NullString
    Presupuesto string
    TotalArea   NullString
    CoveredArea NullString
    Banios      NullString
    Recamaras   NullString
}
```

### UtmDefinition

Códigos UTM detectables en los mensajes de WhatsApp. Cuando el mensaje de un lead contiene uno de estos códigos, los campos UTM de la comunicación se sobreescriben automáticamente.

```go
type UtmDefinition struct {
    Id       int
    Code     string     // Código a buscar en el texto del mensaje
    Source   NullString
    Medium   NullString
    Campaign NullString
    Ad       NullString
    Channel  string     // "ivr"|"inbox"|"whatsapp"|"email"|"flyer"
}
```

### Source

Origen específico de una comunicación. Puede ser genérico (whatsapp, ivr) o estar ligado a una propiedad de un portal.

```go
type Source struct {
    Id         int
    Tipo       string        // "whatsapp", "ivr", "property", etc.
    PropertyId sql.NullInt16 // Referencia a Property si Tipo="property"
}
```

---

## Esquema de Base de Datos

| Tabla | Descripción |
|-------|-------------|
| `Asesor` | Asesores de ventas |
| `Leads` | Leads únicos (deduplicados por teléfono) con asesor asignado |
| `Communication` | Cada consulta/contacto recibido (muchas por lead) |
| `Source` | Origen de cada comunicación |
| `Property` | Propiedades inmobiliarias scraped de portales |
| `Utm` | Definiciones de códigos UTM reconocibles |
| `Message` | Texto de mensajes recibidos por WhatsApp |
| `Action` | Historial de acciones ejecutadas por los flujos |
| `Portal` | Metadatos de portales inmobiliarios |
| `PropertyImages` | Imágenes de propiedades |
| `PublishedProperty` | Estado de publicación de propiedades |

**Relaciones clave:**
```
Communication → Lead → Asesor
Communication → Source → Property (si Source.Tipo = "property")
Communication → Message (si tiene mensaje de WhatsApp)
Lead → Action (última acción ejecutada → define el próximo flujo)
```

---

## Flujos (Flows)

El motor de flujos controla **qué mensajes enviar, cuándo y bajo qué condiciones**. Permite crear secuencias conversacionales automatizadas con delays configurables.

### Estructura de un Flujo

Los flujos se definen en `api/actions.json`:

```json
{
  "Main": "uuid-del-flujo-principal",
  "Flows": {
    "uuid-flujo-A": {
      "Name": "Bienvenida lead nuevo",
      "Rules": [
        {
          "condition": {
            "is_new": true,
            "fuente": "whatsapp"
          },
          "actions": [
            {
              "action": "wpp.message",
              "interval": "0s",
              "params": { "text": "¡Hola {{.Nombre}}! Gracias por contactarnos." },
              "on_response": null
            },
            {
              "action": "wpp.template",
              "interval": "5m",
              "params": {
                "name": "info_propiedad",
                "language": { "code": "es_MX" },
                "components": [
                  {
                    "type": "body",
                    "parameters": [
                      { "type": "text", "text": "{{.Propiedad.Titulo}}" },
                      { "type": "text", "text": "{{.Propiedad.Precio}}" }
                    ]
                  }
                ]
              },
              "on_response": "uuid-flujo-B"
            }
          ]
        }
      ]
    }
  }
}
```

### Condiciones disponibles

Todos los campos son opcionales. Los que se especifiquen deben cumplirse simultáneamente (AND lógico).

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `is_new` | bool | `true` = lead nuevo, `false` = duplicado |
| `fuente` | string | Origen: `"whatsapp"`, `"inmuebles24"`, etc. |
| `telefono` | string | Teléfono exacto del lead |
| `asesor_phone` | string | Teléfono del asesor asignado |
| `asesor_name` | string | Nombre del asesor |
| `nombre` | string | Nombre del lead |
| `message` | string | Texto contenido en el mensaje |
| `utm_source` | string | Valor de utm_source |
| `utm_medium` | string | Valor de utm_medium |
| `utm_campaign` | string | Valor de utm_campaign |
| `fecha_from` | time | Fecha mínima del lead |
| `fecha_to` | time | Fecha máxima del lead |

### Acciones disponibles

| Acción | Descripción | Parámetros |
|--------|-------------|-----------|
| `wpp.message` | Envía mensaje de texto plano por WhatsApp | `text` (soporta template Go) |
| `wpp.template` | Envía template de WhatsApp aprobado por Meta | `name`, `language`, `components` |
| `wpp.media` | Envía imagen o video | `image_id` o `video_id` (ID de Meta) |
| `wpp.cotizacion` | Genera PDF de cotización vía JotForm y lo envía | — |
| `wpp.send_message_asesor` | Notifica al asesor asignado con datos del lead | — |
| `infobip.save` | Guarda/actualiza comunicación en Infobip | — |
| `pipedrive.save` | Guarda/actualiza comunicación en Pipedrive | — |

### Interpolación de texto

Los parámetros de texto soportan la sintaxis de `text/template` de Go. Variables disponibles (campos de `Communication`):

```
{{.Nombre}}               → Nombre del lead
{{.Telefono}}             → Teléfono del lead
{{.Asesor.Name}}          → Nombre del asesor
{{.Asesor.Phone}}         → Teléfono del asesor
{{.Propiedad.Titulo}}     → Título de la propiedad
{{.Propiedad.Precio}}     → Precio de la propiedad
{{.Propiedad.Ubicacion}}  → Ubicación de la propiedad
{{.Propiedad.Link}}       → URL de la propiedad
{{.Cotizacion}}           → URL del PDF de cotización
{{.Utm.Source}}           → UTM source
```

### Delays (`interval`)

Define el tiempo de espera antes de ejecutar la acción. Formato de duración de Go:

```
"0s"   → Inmediatamente
"5m"   → 5 minutos después
"2h"   → 2 horas después
"24h"  → 24 horas después
```

### Lógica de ejecución

Cuando llega una comunicación, el sistema ejecuta:

```
1. StoreCommunication()
   ├─ Determina si el lead es nuevo o duplicado
   ├─ Asigna asesor (round-robin si es nuevo)
   ├─ Detecta códigos UTM en el mensaje recibido
   └─ Guarda en base de datos

2. runAction()
   ├─ Consulta la última Action registrada para este teléfono
   │
   ├─ ¿Tiene campo on_response definido?
   │   └─ SÍ → Ejecuta el flujo apuntado por on_response
   │
   └─ NO → Ejecuta el flujo principal (Main)

3. Ejecución del flujo:
   ├─ Recorre Rules en orden
   ├─ Para cada Rule: evalúa si condition coincide
   └─ Si coincide: programa cada Action tras su delay

4. Al ejecutar cada acción:
   └─ Guarda registro en tabla Action:
      { name, flujo, teléfono, on_response }
      → Permite que la próxima respuesta del lead active el flujo correcto
```

### Flujos conversacionales (on_response)

El campo `on_response` convierte secuencias lineales en conversaciones ramificadas:

```
Sistema envía acción con on_response = "uuid-flujo-B"
                 ↓
         Lead responde por WhatsApp
                 ↓
         runAction() consulta última Action del lead
         → tiene on_response = "uuid-flujo-B"
                 ↓
         Ejecuta "uuid-flujo-B" (en lugar del flujo principal)
```

### Broadcast

Ejecuta un flujo sobre un conjunto de leads filtrados:

```
POST /broadcast
{
    "uuid": "uuid-del-flujo",
    "condition": {
        "fuente": "whatsapp",
        "is_new": true,
        "asesor_phone": "521234567890"
    }
}
```

El sistema:
1. Busca leads distintos que coincidan con `condition`
2. Lanza el flujo para cada uno en goroutines separadas
3. Agrega 100ms de delay entre lanzamientos para respetar rate limits de WhatsApp

---

## Integración WhatsApp

### Envío de mensajes

Usa la WhatsApp Cloud API de Meta:

```
POST https://graph.facebook.com/v17.0/{numberId}/messages
Authorization: Bearer {accessToken}
Content-Type: application/json
```

**Mensaje de texto:**
```json
{
  "messaging_product": "whatsapp",
  "recipient_type": "individual",
  "to": "+521234567890",
  "type": "text",
  "text": { "body": "Hola, ¿en qué te puedo ayudar?" }
}
```

**Template con parámetros:**
```json
{
  "messaging_product": "whatsapp",
  "to": "+521234567890",
  "type": "template",
  "template": {
    "name": "nombre_template",
    "language": { "code": "es_MX" },
    "components": [
      {
        "type": "body",
        "parameters": [
          { "type": "text", "text": "valor1" },
          { "type": "image", "image": { "id": "media-id-meta" } }
        ]
      }
    ]
  }
}
```

**Notificación al asesor** (acción `wpp.send_message_asesor`):

Usa el template `info_asesor_2` e incluye automáticamente:
- Nombre, fuente, teléfono, email del lead
- Indicador de lead nuevo o duplicado
- Datos de propiedad: título, precio, ubicación, URL
- Criterios de búsqueda: presupuesto, área, baños, recámaras
- Datos UTM de tracking

### Recepción de mensajes (Webhook)

Meta envía notificaciones a `POST /webhooks` cuando un usuario escribe al número de WhatsApp Business.

**Verificación inicial** (Meta la realiza una vez al configurar el webhook):
```
GET /webhooks?hub.mode=subscribe&hub.verify_token=TOKEN&hub.challenge=CHALLENGE
→ Responde con el valor de challenge si el token es válido
```

**Tipos de mensaje procesados:**

| Tipo | Cómo se extrae el texto |
|------|------------------------|
| `text` | `message.text.body` |
| `interactive` (botón) | `message.interactive.button_reply.title` |
| `interactive` (lista) | `message.interactive.list_reply.title` |

**Pipeline del webhook:**
```
Meta → POST /webhooks
           ↓
     ReciveNotification()
     Parsea el JSON
           ↓
     Canal interno: entries <- entry
           ↓
     ConsumeEntries() [goroutine permanente en background]
     Entry2Communication()  → crea struct Communication
           ↓
     CommunicationService.NewCommunication()
           ↓
     StoreCommunication() + runAction()
```

---

## API REST

### Comunicaciones

| Método | Ruta | Descripción |
|--------|------|-------------|
| `POST` | `/communication` | Crear comunicación individual |
| `GET` | `/communication` | Listar con filtros y paginación |
| `POST` | `/communication/csv` | Carga masiva desde CSV (máx. 200 registros) |

**Filtros disponibles (GET `/communication`):**

`fecha_from`, `fecha_to`, `asesor_phone`, `asesor_name`, `fuente`, `nombre`, `telefono`, `is_new`, `utm_source`, `utm_medium`, `utm_campaign`, `utm_ad`, `message`, `page`, `limit`

### Asesores

| Método | Ruta | Descripción |
|--------|------|-------------|
| `GET` | `/asesor` | Listar todos los asesores |
| `GET` | `/asesor/{phone}` | Obtener asesor por teléfono |
| `POST` | `/asesor` | Crear asesor |
| `PUT` | `/asesor/{phone}` | Actualizar asesor |
| `DELETE` | `/asesor/{phone}` | Eliminar asesor |
| `PUT` | `/asesor/{phone}/reasign` | Desactivar y redistribuir sus leads |
| `GET` | `/asesor/{phone}/leads` | Ver leads del asesor |

### Flujos

| Método | Ruta | Descripción |
|--------|------|-------------|
| `GET` | `/flows` | Listar todos los flujos |
| `GET` | `/flows/main` | Obtener el flujo principal |
| `GET` | `/flows/{uuid}` | Obtener flujo por UUID |
| `POST` | `/flows` | Crear flujo |
| `PUT` | `/flows/{uuid}` | Actualizar flujo |
| `DELETE` | `/flows/{uuid}` | Eliminar flujo |
| `POST` | `/mainFlow` | Establecer un flujo como principal |
| `POST` | `/broadcast` | Ejecutar flujo sobre leads filtrados |
| `GET` | `/actions` | Listar acciones disponibles con sus schemas JSON |

### UTMs

| Método | Ruta | Descripción |
|--------|------|-------------|
| `GET` | `/utm` | Listar definiciones UTM |
| `POST` | `/utm` | Crear definición UTM |
| `PUT` | `/utm/{id}` | Actualizar definición |
| `DELETE` | `/utm/{id}` | Eliminar definición |

### WhatsApp

| Método | Ruta | Descripción |
|--------|------|-------------|
| `GET` | `/webhooks` | Verificación de webhook (Meta) |
| `POST` | `/webhooks` | Recepción de mensajes entrantes |

### Pipedrive

| Método | Ruta | Descripción |
|--------|------|-------------|
| `GET` | `/pipedrive/auth` | Iniciar flujo OAuth con Pipedrive |
| `GET` | `/pipedrive/callback` | Callback OAuth de Pipedrive |

---

## Procesamiento de Leads

### Flujo completo de ingesta

```
[Portal scraper Python]  [WhatsApp webhook]  [POST /communication]
          └─────────────────────┴──────────────────┘
                                ↓
              CommunicationService.NewCommunication(c)
                                ↓
                    ┌─────────────────────────┐
                    │    StoreCommunication   │
                    │                         │
                    │  1. getOrInsertSource() │
                    │     ├─ whatsapp/ivr:    │
                    │     │  source existente │
                    │     └─ portal:          │
                    │        get/insert       │
                    │        Property+Source  │
                    │                         │
                    │  2. SaveLead()          │
                    │     ├─ ¿existe phone?   │
                    │     │  NO → IsNew=true  │
                    │     │       RR.Next()   │
                    │     │       Insert Lead │
                    │     └─ SÍ → IsNew=false │
                    │           keep asesor   │
                    │           Update Lead   │
                    │                         │
                    │  3. findUtmInMessage()  │
                    │     busca códigos UTM   │
                    │     en el texto recib.  │
                    │                         │
                    │  4. Insert Communication│
                    │  5. Insert Message      │
                    └─────────────────────────┘
                                ↓
                       runAction(communication)
                                ↓
                    ┌─────────────────────────┐
                    │  GetLastAction(phone)   │
                    │  ¿tiene on_response?    │
                    │  SÍ → RunFlow(uuid)     │
                    │  NO  → RunMain()        │
                    └─────────────────────────┘
```

### Round-Robin de Asesores

Cuando llega un lead nuevo, el sistema asigna el siguiente asesor activo en turno:

```
Asesores activos: [Ana, Bruno, Carlos, Diana]

Lead 1 → Ana
Lead 2 → Bruno
Lead 3 → Carlos
Lead 4 → Diana
Lead 5 → Ana   (ciclo)
```

**Reasignación** (`PUT /asesor/{phone}/reasign`):
1. Marca el asesor como `active=false`
2. Recarga el round-robin con los asesores restantes
3. Redistribuye todos sus leads actuales entre los activos

---

## Integraciones

### Pipedrive CRM

Acción: `pipedrive.save`

1. Busca el asesor por email en Pipedrive
2. Busca o crea la persona (lead) por teléfono
3. Busca o crea un Deal para esa persona y asesor
4. Agrega notas con datos de la comunicación

Requiere OAuth. Flujo de autorización: `GET /pipedrive/auth` → login Pipedrive → `GET /pipedrive/callback`

### Infobip

Acción: `infobip.save`

Sincroniza el lead como contacto con atributos personalizados:
- `prop_link`, `prop_precio`, `prop_ubicacion`, `prop_titulo`
- `contacted`, `fuente`, `asesor_name`, `asesor_phone`, `fecha_lead`

### JotForm (Cotizaciones)

Acción: `wpp.cotizacion`

Genera un PDF de cotización via JotForm API usando los datos del lead y lo envía como documento por WhatsApp.

### Reportes diarios

Job configurado con la variable `CRON` que envía a los asesores estadísticas del período:
- Total de comunicaciones
- Desglose por fuente
- Leads nuevos vs. duplicados

---

## Fuentes de Leads

| Fuente | Tipo | Mecanismo |
|--------|------|-----------|
| `whatsapp` | Directo | Webhook WhatsApp Cloud API |
| `ivr` | Directo | Integración telefónica IVR |
| `viewphone` | Directo | Click en teléfono de portal |
| `inmuebles24` | Portal | Scraper Python |
| `lamudi` | Portal | Scraper Python |
| `casasyterrenos` | Portal | Scraper Python |
| `propiedades` | Portal | Scraper Python |
| `csv_file_{fecha}` | Importación | Carga manual vía `POST /communication/csv` |

Los leads de portales incluyen datos de la propiedad de interés que se almacenan en la tabla `Property` y se vinculan a la comunicación via `Source`.

---

## Instalación

### Con Docker

```shell
# Construir el contenedor
docker compose build

# Ejecutar
docker compose up -d

# Detener
docker compose stop

# Ejecutar un scraper específico manualmente
docker compose run app python main.py <portal>
# portales: casasyterrenos | propiedades | lamudi | inmuebles24
```

### Sin Docker

```shell
# Scrapers Python
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
python main.py <portal>

# API Go
cd api
go mod download
go run main.go
```

---

## Testing de WhatsApp

Para validar que los flujos funcionan correctamente antes de activarlos en producción:

1. Ir a la pestaña **Acciones** del panel de Rebora
2. Hacer click en **"Set as main"** en el flow `"TESTING leads nuevos"`
3. Enviar cualquier mensaje de WhatsApp al número de Rebora (se actúa como lead nuevo)
4. Verificar que se ejecuten todas las acciones esperadas (mensajes, imágenes, templates, etc.)
5. Revisar la pestaña **Logs** del panel y buscar errores
6. Volver a establecer como main el flow `"Respuesta"`
