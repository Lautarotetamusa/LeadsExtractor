!#/bin/bash

export $(grep -v '^#' .env| xargs)

#docker compose stop mongo
docker compose up -d mongo
docker compose exec mongo mongosh --username $DB_USER --password $DB_PASS --port 27017

# Obtener todos los logs
# show databeses;
# use mydatabase;
# db.log.find();
