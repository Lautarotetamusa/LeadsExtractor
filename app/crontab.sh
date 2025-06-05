#!/bin/sh
*/10 8-23 * * * python /app/main.py portal lamudi main > /proc/1/fd/1 2> /proc/1/fd/2
*/10 8-23 * * * python /app/main.py portal casasyterrenos main > /proc/1/fd/1 2> /proc/1/fd/2
*/10 8-23 * * * python /app/main.py portal propiedades main > /proc/1/fd/1 2> /proc/1/fd/2
*/10 8-23 * * * python /app/main.py portal inmuebles24 main > /proc/1/fd/1 2> /proc/1/fd/2
