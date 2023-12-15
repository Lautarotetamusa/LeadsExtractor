#!/bin/sh
echo $CRON

# Run supercronic in the background
supercronic crontab.sh &

# Run gunicorn
gunicorn -c gunicorn_config.py app:app
