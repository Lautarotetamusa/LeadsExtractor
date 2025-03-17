# Onedrive images downloader

Este script descarga una foto guardada en onedrive a partir de su url.

## Requisitos
   - pip install -r requirements.txt
   - Variables de entorno actualizadas (APPID y SECRETCLIENT en app creada en AZURE). SECRETCLIENT caduca cada cierto tiempo.

## Uso

1. **Ejecutando el archivo**:

   - Ejecutar main.py
   - Ingresar URL de la foto si es que sigue de esa manera
   - Esperar que abra AZURE y loguearse
   - Esperar a que cargue la url en localhost:3000 (default) y copiarla en su totalidad en la consola (el codigo lo extrae automaticamente el script)

   *Se generarán 2 archivos, uno que guarda el token y otro que guarda el refresh del token para cuando caduque el anterior*

   - Esperar los mensajes de busqueda y descarga de la url y la foto respectivamente
   - Encontrará la foto dentro de la carpeta "photos"

2. **Ejecutando el archivo con los tokens ya generados**:

   - Ejecutar main.py
   - Ingresar URL de la foto si es que sigue de esa manera
   - Esperar a mensajes de busqueda y descaga de la foto