-- +goose Up
-- Elimina la doble semántica de Source (fila fija reusada vs. fila 1-a-1 por
-- Property) moviendo esa información directo a Communication: `fuente` pasa
-- a ser el valor real (whatsapp/ivr/viewphone/inmuebles24/...) en vez del
-- genérico "property", y `property_id` es un FK nullable a Property. Esto no
-- cambia el JSON de POST /communication (Portalia sigue mandando lo mismo).

ALTER TABLE Communication
    ADD COLUMN fuente ENUM('whatsapp', 'ivr', 'viewphone', 'inmuebles24', 'lamudi', 'casasyterrenos', 'propiedades') NULL,
    ADD COLUMN property_id INT NULL;

UPDATE Communication C
    JOIN Source S ON C.source_id = S.id
    LEFT JOIN Property P ON S.property_id = P.id
SET
    C.fuente = IF(S.type = 'property', P.portal, S.type),
    C.property_id = S.property_id;

ALTER TABLE Communication
    MODIFY COLUMN fuente ENUM('whatsapp', 'ivr', 'viewphone', 'inmuebles24', 'lamudi', 'casasyterrenos', 'propiedades') NOT NULL,
    ADD CONSTRAINT chk_comm_fuente_property CHECK (
        (fuente IN ('whatsapp', 'ivr', 'viewphone') AND property_id IS NULL) OR
        (fuente NOT IN ('whatsapp', 'ivr', 'viewphone') AND property_id IS NOT NULL)
    ),
    ADD CONSTRAINT fk_comm_property FOREIGN KEY (property_id) REFERENCES Property(id);

-- Dropear la FK vieja de source_id requiere su nombre real: MySQL/MariaDB lo
-- autogeneran (no lo pusimos explícito en 00001_baseline.sql).
SET @fk_name = (
    SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'Communication'
      AND COLUMN_NAME = 'source_id'
      AND REFERENCED_TABLE_NAME = 'Source'
    LIMIT 1
);
SET @drop_fk_sql = CONCAT('ALTER TABLE Communication DROP FOREIGN KEY ', @fk_name);
PREPARE drop_fk_stmt FROM @drop_fk_sql;
EXECUTE drop_fk_stmt;
DEALLOCATE PREPARE drop_fk_stmt;

ALTER TABLE Communication DROP COLUMN source_id;

DROP TABLE Source;

-- +goose Down
CREATE TABLE Source (
    id INT NOT NULL AUTO_INCREMENT,
    type ENUM("whatsapp", "ivr", "property", "viewphone") NOT NULL,
    property_id INT DEFAULT NULL,

    CHECK (
        (type = 'property' AND property_id IS NOT NULL) OR
        (type IN ('whatsapp', 'ivr', 'viewphone') AND property_id IS NULL)
    ),

    PRIMARY KEY (id),
    FOREIGN KEY (property_id) REFERENCES Property(id)
);
INSERT INTO Source (type) VALUES ("whatsapp"), ("ivr"), ("viewphone");

ALTER TABLE Communication ADD COLUMN source_id INT NULL;

UPDATE Communication C
    JOIN Source S ON S.type = C.fuente
SET C.source_id = S.id
WHERE C.fuente IN ('whatsapp', 'ivr', 'viewphone');

INSERT INTO Source (type, property_id)
    SELECT 'property', C.property_id
    FROM Communication C
    WHERE C.property_id IS NOT NULL
    GROUP BY C.property_id;

UPDATE Communication C
    JOIN Source S ON S.property_id = C.property_id
SET C.source_id = S.id
WHERE C.property_id IS NOT NULL;

ALTER TABLE Communication
    MODIFY COLUMN source_id INT NOT NULL,
    ADD FOREIGN KEY (source_id) REFERENCES Source(id),
    DROP FOREIGN KEY fk_comm_property,
    DROP CONSTRAINT chk_comm_fuente_property,
    DROP COLUMN property_id,
    DROP COLUMN fuente;
