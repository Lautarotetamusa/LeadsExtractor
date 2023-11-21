#!/bin/sh
0 10-18/4 * * * python /app/main.py lamudi > /proc/1/fd/1 2> /proc/1/fd/2
0 10-18/4 * * * python /app/main.py casasyterrenos > /proc/1/fd/1 2> /proc/1/fd/2
0 10-18/4 * * * python /app/main.py propiedades > /proc/1/fd/1 2> /proc/1/fd/2
0 10-18/4 * * * python /app/main.py inmuebles24 > /proc/1/fd/1 2> /proc/1/fd/2