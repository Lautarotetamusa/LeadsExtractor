-- +goose Up
-- Esquema consolidado a partir de db/db.sql + las migraciones sueltas
-- 01 a 15 que vivían en db/ (aplicadas a mano hasta ahora). A partir de acá
-- las migraciones nuevas se agregan con `goose create` en este directorio.

CREATE TABLE Asesor(
    name   VARCHAR(64)  NOT NULL,
    phone  CHAR(16)     NOT NULL,
    email  VARCHAR(128) NOT NULL,
    active BOOLEAN,

    CHECK (phone > 0),

    PRIMARY KEY (phone)
);

CREATE TABLE Leads(
    name   VARCHAR(64) NOT NULL,
    phone  CHAR(16)    NOT NULL,
    email  VARCHAR(64) DEFAULT NULL,
    asesor CHAR(16)    NOT NULL,

    CHECK (phone > 0),
    CHECK (asesor > 0),

    PRIMARY KEY (phone),
    FOREIGN KEY (asesor) REFERENCES Asesor(phone)
);

-- Snapshot de la propiedad de interés de la comunicación (los datos
-- completos de la propiedad y su publicación viven en Portalia).
CREATE TABLE Property(
    id INT NOT NULL AUTO_INCREMENT,
    portal ENUM("inmuebles24", "lamudi", "casasyterrenos", "propiedades") NOT NULL,

    portal_id    VARCHAR(128) DEFAULT NULL,
    title        VARCHAR(256) DEFAULT NULL,
    price        VARCHAR(32)  DEFAULT NULL,
    ubication    VARCHAR(256) DEFAULT NULL,
    url          VARCHAR(256) DEFAULT NULL,
    tipo         VARCHAR(32)  DEFAULT NULL,
    bedrooms     VARCHAR(16)  DEFAULT NULL,
    bathrooms    VARCHAR(16)  DEFAULT NULL,
    total_area   VARCHAR(16)  DEFAULT NULL,
    covered_area VARCHAR(16)  DEFAULT NULL,

    CHECK (portal_id IS NOT NULL AND portal_id != ''),

    PRIMARY KEY (id)
);

CREATE TABLE Source(
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
-- No se siembra una fila de type="property": esas filas se crean dinámicamente
-- por cada propiedad real (ver SourceDBStore.InsertSource) y el CHECK de abajo
-- exige property_id NOT NULL para ese type, así que una fila semilla violaría
-- la constraint (nunca se detectó antes porque MySQL 5.7 no aplica CHECKs).
INSERT INTO Source (type) VALUES ("whatsapp"), ("ivr"), ("viewphone");

CREATE TABLE Communication(
    id INT NOT NULL AUTO_INCREMENT,
    lead_phone CHAR(16) NOT NULL,
    source_id INT NOT NULL,
    url VARCHAR(256) DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    new_lead BOOLEAN NOT NULL DEFAULT true,

    lead_date DATE NOT NULL,
    zones VARCHAR(256) DEFAULT NULL,
    mt2_terrain VARCHAR(32) DEFAULT NULL,
    mt2_builded VARCHAR(32) DEFAULT NULL,
    baths VARCHAR(32) DEFAULT NULL,
    rooms VARCHAR(32) DEFAULT NULL,

    utm_source   VARCHAR(256) DEFAULT NULL,
    utm_medium   VARCHAR(256) DEFAULT NULL,
    utm_campaign VARCHAR(256) DEFAULT NULL,
    utm_ad       VARCHAR(256) DEFAULT NULL,
    utm_channel  VARCHAR(256) DEFAULT NULL,

    CHECK (lead_phone > 0),

    PRIMARY KEY (id),
    FOREIGN KEY (lead_phone) REFERENCES Leads(phone),
    FOREIGN KEY (source_id) REFERENCES Source(id)
);

CREATE TABLE Utm (
    id INT NOT NULL AUTO_INCREMENT,
    code VARCHAR(64) NOT NULL,
    utm_source   VARCHAR(256) DEFAULT NULL,
    utm_medium   VARCHAR(256) DEFAULT NULL,
    utm_campaign VARCHAR(256) DEFAULT NULL,
    utm_ad       VARCHAR(256) DEFAULT NULL,
    utm_channel  ENUM('ivr', 'inbox', 'whatsapp', 'email', 'flyer'),

    PRIMARY KEY (id),
    UNIQUE (code)
);

-- Texto del mensaje de WhatsApp recibido en una comunicación.
CREATE TABLE Message (
    id INT NOT NULL AUTO_INCREMENT,
    id_communication INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    text TEXT NOT NULL,
    wamid VARCHAR(64) DEFAULT NULL,

    CHECK (text != ""),

    PRIMARY KEY (id),
    FOREIGN KEY (id_communication) REFERENCES Communication(id)
);

-- Cola de mensajes programados del módulo `messenger` (cmd/messenger).
CREATE TABLE Messages (
    id           BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
    phone        CHAR(16)    NOT NULL,
    type         VARCHAR(32) NOT NULL DEFAULT 'text',
    content      TEXT        NOT NULL,
    outgoing     BOOLEAN     NOT NULL DEFAULT FALSE,
    scheduled_at DATETIME    NOT NULL,
    sended_at    DATETIME    DEFAULT NULL,
    on_response  CHAR(37)    DEFAULT NULL,

    INDEX idx_phone   (phone),
    INDEX idx_pending (outgoing, scheduled_at, sended_at)
);

-- Historial de acciones de flow ejecutadas (define qué flow sigue si el lead responde).
CREATE TABLE Action (
    name        VARCHAR(256) NOT NULL,
    nro         INT          NOT NULL,
    flow_uuid   CHAR(37)     NOT NULL,
    sended_at   TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    lead_phone  CHAR(16)     NOT NULL,
    text        TEXT         NOT NULL DEFAULT '',
    on_response CHAR(37)     DEFAULT NULL,

    FOREIGN KEY (lead_phone) REFERENCES Leads(phone)
);

-- +goose StatementBegin
CREATE PROCEDURE getCommunications (IN date_from DATETIME, IN is_new BOOLEAN)
    BEGIN
        SELECT
            C.created_at,
            C.lead_date,
            C.new_lead,
            A.name as "asesor.name", A.phone as "asesor.phone", A.email as "asesor.email",
            IF(S.type = "property", P.portal, S.type) as "fuente",
            L.name, C.url, L.phone, L.email,
                IFNULL(P.portal_id, "") as "propiedad.portal_id",
                IFNULL(P.title, "") as "propiedad.title",
                IFNULL(P.price, "") as "propiedad.price",
                IFNULL(P.ubication, "") as "propiedad.ubication",
                IFNULL(P.url, "") as "propiedad.url",
                IFNULL(P.tipo, "") as "propiedad.tipo",
            C.zones as "busquedas.zones", C.mt2_terrain as "busquedas.mt2_terrain", C.mt2_builded as "busquedas.mt2_builded", C.baths as "busquedas.baths", C.rooms as "busquedas.rooms"
        FROM Communication C
        INNER JOIN Leads L
            ON C.lead_phone = L.phone
        INNER JOIN Source S
            ON C.source_id = S.id
        INNER JOIN Asesor A
            ON L.asesor = A.phone
        LEFT JOIN Property P
            ON S.property_id = P.id
        WHERE date_from IS NULL OR C.created_at > date_from
        AND     is_new IS NULL OR C.new_lead = is_new
        ORDER BY C.id DESC;
    END;
-- +goose StatementEnd

INSERT INTO Asesor (name, phone, email, active) VALUES
    ("Brenda Díaz", "5213313420733", "brenda.diaz@rebora.com.mx", False),
    ("Aldo Salcido", "5213322563353", "aldo.salcido@rebora.com.mx", False),
    ("Onder Sotomayor", "5213318940377", "onder.sotomayor@rebora.com.mx", True),
    ("Diego Rubio", "5213317186543", "diego.rubio@rebora.com.mx", False),
    ("Maggie Escobedo", "5213314299454", "maggie.escobedo@rebora.com.mx", False);

-- +goose Down
DROP PROCEDURE IF EXISTS getCommunications;
DROP TABLE IF EXISTS Action;
DROP TABLE IF EXISTS Messages;
DROP TABLE IF EXISTS Message;
DROP TABLE IF EXISTS Utm;
DROP TABLE IF EXISTS Communication;
DROP TABLE IF EXISTS Source;
DROP TABLE IF EXISTS Property;
DROP TABLE IF EXISTS Leads;
DROP TABLE IF EXISTS Asesor;
