#!/bin/sh

# Run supercronic in the background
supercronic crontab.sh &

# Run gunicorn
#debug: 
#gunicorn --log-level=debug -c gunicorn.py server:app
gunicorn -c gunicorn.py server:app
