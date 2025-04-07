## Generar el archivo HTML
Para generar el archivo html apartir del pdf que envió el diseñador use un programa que se llama
[pdf2htmlEX](https://github.com/coolwanglu/pdf2htmlEX)

Genero un html muy parecido al pdf original, se puede hacer desde una web o instalarlo local.

Para usarlo local lo más fácil es instalarse la imagen de docker:
https://github.com/sergiomtzlosa/docker-pdf2htmlex

- 2.
Despues fui mirando en código html y reemplazando a mano todos los valores que tienen que ser calculados
por ejemplo: {{valor_total}}

## Generar de vuelta un pdf
El html creado en teoría permite descargarlo como pdf de vuelta sin ningun problema.
En firefox directamente no me sale la opción para descargarlo como pdf. En chrome lo pude descargar bien cambiando un poco los ajustes.

En Chrome, para que salga bien pongo de opciones de configuracion las siguientes:
Tamaño de pape: A4.
Margenes: Ninguno
Escala: Predeterminado.
