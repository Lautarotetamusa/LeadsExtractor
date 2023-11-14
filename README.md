# Estado del proyecto

## [Inmuebles24](src/inmuebles24/inmuebles24.md)

- [X] Fase de investigacion y testeo
- [ ] Fase de desarrollo

### To do:

- [X] Enviar mensajes
- [ ] inicio de sesion
- [ ] Detectar token vencido y volver a loggearnos

## [Lamudi]()

- [X] Fase de investigacion y testeo
- [X] Fase de desarrollo
- [X] Primera corrida

### To do:

- [X] Detectar token vencido y volver a loggearnos

## [Casas y Terrenos]()

- [X] Fase de investigacion
- [X] Fase de desarrollo
- [X] Primera corrida

### To do:

- [X] Detectar token vencido y volver a loggearnos

## [Propiedades.com]()

- [X] Fase de investigacion
- [X] Fase de desarrollo
- [X] Primera corrida

### To do:

- [ ] Detectar token vencido y volver a loggearnos

# El objecto lead, estructura

Todos los objectos que representan la informacion de un lead están estandarizados con el siguiente formato:

```json
{
    "fuente": "string",
    "id": "string",
    "fecha": "string",
    "nombre": "string",
    "link": "string",
    "telefono": "string",
    "email": "string",
    "propiedad": {
        "id": "string",
        "titulo": "string",
        "link": "string",
        "precio": "string",
        "ubicacion": "string",
        "tipo": "string",
    },
    "busquedas": {
        "zonas": "string", //zonas separadas por comas
        "tipo": "string", 
        "presupuesto": "string",
        "cantidad_anuncios": "string",
        "contactos": "string", //cantidad de contactos que realizo
        "inicio_busqueda": "string", //min - max
        "total_area": "string", //min - max
        "covered_area": "string", //min - max
        "banios": "string", //min - max
        "recamaras": "string", //min - max
    }
}
```

# Mapping

En el archivo `mapping.json` configuraremos qué columnas de nuestro archivo google sheets están conectadas con qué campos de nuestros objectos lead.

mapping.json:

```json
{
    "Fecha": "fecha", 
    "Fuente": "fuente", 
    "Nombre del lead": "nombre", 
    "Link a consulta": "link", 
    "Telefono 1": "telefono", 
    "Telefono 2": "", 
    "Email": "email", 
    "Titulo": "propiedad.titulo", 
    "Link": "propiedad.link", 
    "Precio": "propiedad.precio", 
    "Ubicación": "propiedad.ubicacion", 
    "Tipo": "propiedad.tipo", 
    "Zonas": "busquedas.zonas",
    "Mt2 terreno": "busquedas.total_area",
    "Mt2 construcción": "busquedas.covered_area", 
    "Baños": "busquedas.banios", 
    "Recámaras": "busquedas.recamaras", 
    "Presupuesto": "busquedas.presupuesto", 
    "Anuncios vistos": "busquedas.cantidad_anuncios", 
    "# anuncios contactados": "busquedas.contactos", 
    "Hace cuanto empezo a buscar": "busquedas.inicio_busqueda",
    "Mensaje": "message"
}
```

Las claves a la derecha inidican el nombre de la columna en el google sheet.
Los valores a la izquierda indican que campo del objecto lead utilizaremos, un punto (.) indica que usamos una propiedad dentro de otra, por ejemplo busquedas.presupuesto indica que dentro de la 'busquedas' tomamos la propiedad 'presupuesto'.

# Programacion de los scrapers

Los scrapers estan por defecto programados cada 4 horas entre las 10 y las 18hs.
Esto se puede cambiar en el archivo crontab utilizando la sintaxis de cron

[Aqui podemos generar crons para las horas que querramos](https://crontab.guru/)

crontab:

```shell
0 10-18/4 * * * python /app/inmuebles24/inmuebles24.py > /proc/1/fd/1 2> /proc/1/fd/2
```

Cambiando `0 10-18/4 * * *` podemos programar los scrapers cuando querramos.

# Instalacion

## Instalar docker

## Buildear el contenedor

   `docker compose build`

## Ejecutar el contenedor

 `docker compose up -d`

## Detener el contenedor

 `docker compose stop`

## Ejecutar algun scraper en particular sin esperar

 `docker compose run app python main.py <PORTAL>`
 PORTALS:

- casasyterrenos
- propiedades
- lamudi

## Instalacion sin docker

  `python -m venv .venv`

  `source .venv/bin/activate`

  `pip install -r requeriments.txt`

### Ejecutar algun scraper en particular sin esperar

   `python main.py <PORTAL>`
    PORTALS:
    - casasyterrenos
    - propiedades
    - lamudi
