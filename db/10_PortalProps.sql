SET FOREIGN_KEY_CHECKS=0; -- to disable them
DROP TABLE IF EXISTS PortalProp;
CREATE TABLE IF NOT EXISTS PortalProp(
    id INT NOT NULL AUTO_INCREMENT,

    title VARCHAR(256) NOT NULL,
    price VARCHAR(32) NOT NULL,
    currency CHAR(5) NOT NULL,
    description VARCHAR(4096) NOT NULL, 
    type ENUM("house", "apartment") NOT NULL,
    antiquity INT NOT NULL,
    parkinglots INT NOT NULL DEFAULT 0,
    bathrooms INT NOT NULL DEFAULT 0,
    half_bathrooms INT NOT NULL DEFAULT 0,
    rooms INT NOT NULL DEFAULT 0,
    operation_type ENUM("sell", "rent") NOT NULL,
    m2_total INT,
    m2_covered INT,
    video_url VARCHAR(256) DEFAULT NULL,
    virtual_route VARCHAR(256) DEFAULT NULL,

    zone VARCHAR(256) NOT NULL,

    /* Ubication fields */
    address VARCHAR(256) NOT NULL, /*Full google valid address*/
    lat FLOAT NOT NULL,
    lng FLOAT NOT NULL,

    CHECK (title <> ""),
    CHECK (zone <> ""),
    CHECK (price <> ""),
    CHECK (currency <> ""),
    CHECK (description <> ""),
    CHECK (address <> ""),

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY(id)
);

DROP TABLE IF EXISTS PropertyImages;
CREATE TABLE IF NOT EXISTS PropertyImages (
    id INT NOT NULL AUTO_INCREMENT,

    property_id INT NOT NULL,
    url VARCHAR(512) NOT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (property_id) REFERENCES PortalProp(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS Portal;
CREATE TABLE IF NOT EXISTS Portal (
    name    VARCHAR(64) NOT NULL,
    url     VARCHAR(256) NOT NULL,
    base_url VARCHAR(64) NOT NULL,

    CHECK (name <> ""),
    CHECK (url <> ""),
    check (base_url <> ""),

    PRIMARY KEY(name)
);

INSERT INTO Portal (name, url, base_url) VALUES
("lamudi", "https://lamudi.com.mx",  "https://www.lamudi.com.mx/detalle/"),
("inmuebles24", "https://inmuebles24.com", " - "),
("casasyterrenos", "https://casasyterrenos.com",  "https://www.casasyterrenos.com/propiedad/-"),
("propiedades", "https://propiedades.com", "https://propiedades.com/inmuebles/-");

DROP TABLE IF EXISTS PublishedProperty;
CREATE TABLE IF NOT EXISTS PublishedProperty (
    property_id INT NOT NULL,

    -- The internal id of the property in the portal
    publication_id VARCHAR(64) DEFAULT NULL,

    url     VARCHAR(256) DEFAULT NULL,
    status  ENUM("in_progress", "in_queue", "published", "failed") DEFAULT "in_progress" NOT NULL,
    portal  VARCHAR(64) NOT NULL,

    plan    ENUM("simple", "highlighted", "super") DEFAULT "simple" NOT NULL,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CHECK (url <> ""),
    CHECK (status <> ""),
    CHECK (portal <> ""),

    PRIMARY KEY (portal, property_id),
    FOREIGN KEY (portal) REFERENCES Portal(name),
    FOREIGN KEY (property_id) REFERENCES PortalProp(id) ON DELETE CASCADE
);
SET FOREIGN_KEY_CHECKS=1; -- to re-enable them
