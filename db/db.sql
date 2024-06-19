CREATE USER IF NOT EXISTS 'teti'@'%' IDENTIFIED BY 'Lautaro123.';
GRANT ALL PRIVILEGES ON *.* TO 'teti'@'%' WITH GRANT OPTION;

DROP DATABASE IF EXISTS LeadsExtractor;
CREATE DATABASE IF NOT EXISTS LeadsExtractor;
USE LeadsExtractor;

DROP TABLE IF EXISTS Asesor;
CREATE TABLE IF NOT EXISTS Asesor(
    name  VARCHAR(64) NOT NULL,
    phone CHAR(16) NOT NULL,
    active BOOLEAN,

    CHECK (phone > 0),

    PRIMARY KEY (phone)
);

DROP TABLE IF EXISTS Leads;
CREATE TABLE IF NOT EXISTS Leads(
    name  VARCHAR(64) NOT NULL,
    phone CHAR(16) NOT NULL,
    email VARCHAR(64) DEFAULT NULL,
    asesor CHAR(16) NOT NULL,
    cotizacion VARCHAR(256) DEFAULT "",

    CHECK (phone > 0),
    CHECK (asesor > 0),

    PRIMARY KEY (phone),
    FOREIGN KEY (asesor) REFERENCES Asesor(phone)
);

DROP TABLE IF EXISTS Property;
CREATE TABLE IF NOT EXISTS Property(
    id INT NOT NULL AUTO_INCREMENT,
    portal ENUM("inmuebles24", "lamudi", "casasyterrenos", "propiedades") NOT NULL,

    portal_id VARCHAR(128) DEFAULT NULL,
    title VARCHAR(256) DEFAULT NULL,
    price VARCHAR(32) DEFAULT NULL,
    ubication VARCHAR(256) DEFAULT NULL,
    url VARCHAR(256) DEFAULT NULL,
    tipo VARCHAR(32) DEFAULT NULL,
    
    CHECK (portal_id IS NOT NULL AND portal_id != ''),

    PRIMARY KEY (id)
);

DROP TABLE IF EXISTS Source;
CREATE TABLE IF NOT EXISTS Source(
    id INT NOT NULL AUTO_INCREMENT,
    type ENUM("whatsapp", "ivr", "property") NOT NULL, 
    property_id INT DEFAULT NULL,

    CHECK ( 
        (type = 'property' AND property_id IS NOT NULL) OR
        (type IN ('whatsapp', 'ivr', 'viewphone') AND property_id IS NULL)
    ),
    
    PRIMARY KEY (id),
    FOREIGN KEY (property_id) REFERENCES Property(id)
);
INSERT INTO Source (type) VALUES ("whatsapp"), ("ivr"), ("property");

DROP TABLE IF EXISTS Communication;
CREATE TABLE IF NOT EXISTS Communication(
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
    rooms VARCHAR(32) DEFAULT NUll,

    CHECK (lead_phone > 0),

    PRIMARY KEY (id),
    FOREIGN KEY (lead_phone) REFERENCES Leads(phone),
    FOREIGN KEY (source_id) REFERENCES Source(id)
);

DELIMITER //
DROP PROCEDURE IF EXISTS getCommunications;
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
//
DELIMITER ; 

CALL getCommunications('2024-01-01', null);


INSERT INTO Asesor (name, phone, active) VALUES 
    ("Brenda Díaz", "5213313420733", False),
    ("Aldo Salcido", "5213322563353", False),
    ("Onder Sotomayor", "5213318940377", True),
    ("Diego Rubio", "5213317186543", False),
    ("Maggie Escobedo", "5213314299454", False);
