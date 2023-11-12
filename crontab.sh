#!/bin/sh
0 10-18/4 * * * python /app/lamudi/lamudi.py > /proc/1/fd/1 2> /proc/1/fd/2
0 10-18/4 * * * python /app/casas_y_terrenos/casas_y_terrenos.py > /proc/1/fd/1 2> /proc/1/fd/2
0 10-18/4 * * * python /app/propiedades_com/propiedades.py > /proc/1/fd/1 2> /proc/1/fd/2
#0 10-18/4 * * * python /app/inmuebles24/inmuebles24.py > /proc/1/fd/1 2> /proc/1/fd/2
