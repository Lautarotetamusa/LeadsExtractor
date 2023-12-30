#!/bin/sh
0 9-21/1 * * * python /app/main.py lamudi > /proc/1/fd/1 2> /proc/1/fd/2
0 9-21/1 * * * python /app/main.py casasyterrenos > /proc/1/fd/1 2> /proc/1/fd/2
0 9-21/1 * * * python /app/main.py propiedades > /proc/1/fd/1 2> /proc/1/fd/2
0 9-21/1 * * * python /app/main.py inmuebles24 > /proc/1/fd/1 2> /proc/1/fd/2

#Para la primer corrida de gmail
#Todos los dias a las 9am
#0 9 * * * python /app/gmail_first_run.py > /proc/1/fd/1 2> /proc/1/fd/2
