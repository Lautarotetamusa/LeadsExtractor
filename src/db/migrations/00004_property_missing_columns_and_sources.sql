-- +goose Up
-- goose_db_version marca 00001_baseline.sql como aplicada en todos los
-- ambientes, pero en producción la tabla Property real nunca tuvo
-- bedrooms/bathrooms/total_area/covered_area (el INSERT de
-- store/property.go las usa desde siempre) — probablemente baseline.sql se
-- escribió como consolidación después de que production ya estaba en pie,
-- sin que ese ALTER se aplicara a mano en su momento. Esto hacía fallar
-- (y perder) cualquier lead con una property nueva (no cacheada), porque
-- InsertProperty solo se llama para properties que GetProperty no encontró.
-- IF NOT EXISTS la hace segura de correr también en ambientes que sí las
-- tienen (ej. una DB local migrada desde cero con baseline.sql).
ALTER TABLE Property
    ADD COLUMN IF NOT EXISTS bedrooms     VARCHAR(16) DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS bathrooms    VARCHAR(16) DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS total_area   VARCHAR(16) DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS covered_area VARCHAR(16) DEFAULT NULL;

-- easybroker y wordpress se agregaron como fuentes válidas en
-- store.validSources (Go) pero nunca al ENUM real de la columna: sin este
-- ALTER, cualquier property de esas dos fuentes falla el INSERT igual que
-- el bug de arriba.
ALTER TABLE Property
    MODIFY COLUMN portal ENUM('inmuebles24', 'lamudi', 'casasyterrenos', 'propiedades', 'easybroker', 'wordpress') NOT NULL;

-- +goose Down
ALTER TABLE Property
    MODIFY COLUMN portal ENUM('inmuebles24', 'lamudi', 'casasyterrenos', 'propiedades') NOT NULL;

ALTER TABLE Property
    DROP COLUMN IF EXISTS bedrooms,
    DROP COLUMN IF EXISTS bathrooms,
    DROP COLUMN IF EXISTS total_area,
    DROP COLUMN IF EXISTS covered_area;
