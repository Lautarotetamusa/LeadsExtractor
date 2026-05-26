CREATE TABLE IF NOT EXISTS Action (
    name        VARCHAR(256)  NOT NULL,
    nro         INT           NOT NULL,
    flow_uuid   CHAR(37)      NOT NULL,
    sended_at   TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    lead_phone  CHAR(16)      NOT NULL,
    text        TEXT          NOT NULL DEFAULT '',
    on_response CHAR(37)      DEFAULT NULL,

    FOREIGN KEY (lead_phone) REFERENCES Leads(phone)
);

-- Para DBs que ya tenían la tabla sin la columna text
ALTER TABLE Action ADD COLUMN IF NOT EXISTS text TEXT NOT NULL DEFAULT '';
