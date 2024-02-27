# Objetivo del proyecto

El objetivo de este proyecto es extraer la información de [Leads](leads) de diferentes portales inmobiliarios.
Para cada portal extraeremos toda la información de los diferentes leads y la volcaremos en un archivo de google sheets. Para cada lead, en los portales donde sea posible, se les enviará un mensaje generado automaticamente apartir de los datos de ese lead y un formato predifinido. Este proyecto permitirá acelerar el proceso de seguimiento de los leads automatizando las actividades mas repetitivas através de *Digital workers*.

Los portales con los que trabajaremos por ahora serán:

* [inmuebles24](https://inmuebles24.com)
* [lamudi](https://www.lamudi.com.mx/)
* [Casas y terrenos](https://www.casasyterrenos.com/)
* [Propiedades.com](https://propiedades.com/)

# Estado del proyecto

## [Inmuebles24](src/inmuebles24/inmuebles24.md)

- [X] Fase de investigacion y testeo
- [X] Fase de desarrollo

### To do:

- [X] Enviar mensajes
- [X] inicio de sesion
- [X] Detectar token vencido y volver a loggearnos

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

# Ejecutar scripts
Para ejecutar los scripts deberemos primero asegurarnos de tener instalados todos los requerimientos del archivo
`requirements.txt`. Luego ejecutaremos el siguiente comando:

```shell
python -m scripts.{script}
```

Es importante mencionar que no pondremos la extension .py al ejecutar el comando

# El archivo .env

El archivo .env será donde se configurarán todas las claves necesarias para correr el proyecto.

```
CASASYTERRENOS_USERNAME=""
CASASYTERRENOS_PASSWORD=""

INMUEBLES24_USERNAME=""
INMUEBLES24_PASSWORD=""

LAMUDI_USERNAME=""
LAMUDI_PASSWORD=""

PROPIEDADES_USERNAME=""
PROPIEDADES_PASSWORD=""

SHEET_ID=""

ZENROWS_APIKEY=""

CRON="0 10-18/4 * * *"
```

`<PORTAL>_USERNAME`: Nombre de usuario o correo de la cuenta del `<PORTAL>` \
`<PORTAL>_PASSWORD`: Contraseña para el username del `<PORTAL>` \
`SHEET_ID`: \
    Es el id del archivo de google sheet donde se guardará la información de los leads. \
    El archivo deberá tener en la primera fila (headers) los mismos campos indicados en el archivo [mapping.json](#Mapping)
`ZENROWS_APIKEY`: Es la clave de [ZenRows](https://www.zenrows.com). \

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

# Posibles mejoras
  * Diseñar y desarrollar una base de datos, esto permitiría conectar los datos de los diferentes portales sin repetir información, lo que permitiría un mejor manejo de la información.
  * Nuevos portales.
