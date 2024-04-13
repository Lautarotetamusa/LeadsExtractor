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
    
    CHECK (portal_id IS NOT NULL AND portal_id != '')

    PRIMARY KEY (id)
);

DROP TABLE IF EXISTS Source;
CREATE TABLE IF NOT EXISTS Source(
    id INT NOT NULL AUTO_INCREMENT,
    type ENUM("whatsapp", "ivr", "property") NOT NULL, 
    property_id INT DEFAULT NULL,
    
    PRIMARY KEY (id),
    FOREIGN KEY (property_id) REFERENCES Property(id)
);
INSERT INTO Source (type) VALUES ("whatsapp"), ("property");

DROP TABLE IF EXISTS Communication;
CREATE TABLE IF NOT EXISTS Communication(
    id INT NOT NULL AUTO_INCREMENT,
    lead_phone CHAR(16) NOT NULL,
    source_id INT NOT NULL,
    url VARCHAR(256) DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

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

DROP PROCEDURE IF EXISTS communicationList;
DELIMITER //
CREATE PROCEDURE communicationList ()
    BEGIN
        SELECT 
            C.created_at as "Fecha extraccion", C.lead_date as "Fecha lead", A.name as "Asesor asignado", S.type, P.portal, L.name, C.url, L.phone, L.email,
            P.*,
            C.zones, C.mt2_terrain, C.mt2_builded, C.baths, C.rooms
        FROM Communication C
        INNER JOIN Leads L 
            ON C.lead_phone = L.phone
        INNER JOIN Source S
            ON C.source_id = S.id
        INNER JOIN Asesor A
            ON L.asesor = A.phone
        LEFT JOIN Property P
            ON S.property_id = P.id
        ORDER BY C.id DESC;
    END;
//
DELIMITER ;

INSERT INTO Asesor (name, phone, active) VALUES 
    ("Brenda Díaz", "5213313420733", False),
    ("Aldo Salcido", "5213322563353", False),
    ("Onder Sotomayor", "5213318940377", True),
    ("Diego Rubio", "5213317186543", False),
    ("Maggie Escobedo", "5213314299454", False);
