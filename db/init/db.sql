-- Solo crea el usuario y la base vacía. El esquema lo aplica goose
-- automáticamente al arrancar el servicio (ver src/db/migrations y
-- src/store/migrate.go).
CREATE USER IF NOT EXISTS 'teti'@'%' IDENTIFIED BY 'Lautaro123.';
GRANT ALL PRIVILEGES ON *.* TO 'teti'@'%' WITH GRANT OPTION;

CREATE DATABASE IF NOT EXISTS LeadsExtractor;
