#!/bin/sh
echo $CRON

# Run supercronic in the background
supercronic crontab.sh &

# Run gunicorn
gunicorn --log-level=debug -c gunicorn.py server:app
