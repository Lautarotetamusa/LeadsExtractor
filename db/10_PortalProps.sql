DROP TABLE IF EXISTS PortalProp;
CREATE TABLE IF NOT EXISTS PortalProp(
    id INT NOT NULL AUTO_INCREMENT,

    title VARCHAR(256) NOT NULL,
    price VARCHAR(32) NOT NULL,
    currency CHAR(5) NOT NULL,
    description VARCHAR(512) NOT NULL, 
    type VARCHAR(32) NOT NULL,
    antiquity INT NOT NULL,
    parkinglots INT DEFAULT NULL,
    bathrooms INT DEFAULT NULL,
    half_bathrooms INT DEFAULT NULL,
    rooms INT DEFAULT NULL,
    operation_type ENUM("sale", "rent") NOT NULL,
    m2_total INT,
    m2_covered INT,
    video_url VARCHAR(256) DEFAULT NULL,
    virtual_route VARCHAR(256) DEFAULT NULL,

    /* Ubication fields */
    state VARCHAR(128) NOT NULL,
    municipality VARCHAR(128) NOT NULL,
    colony VARCHAR(128) NOT NULL,
    neighborhood VARCHAR(128) DEFAULT NULL,
    street VARCHAR(256) NOT NULL,
    number VARCHAR(32) NOT NULL,
    zip_code VARCHAR(32) NOT NULL,

    CHECK (title <> ""),
    CHECK (price <> ""),
    CHECK (currency <> ""),
    CHECK (description <> ""),
    CHECK (type <> ""),
    CHECK (state <> ""),
    CHECK (municipality <> ""),
    CHECK (colony <> ""),
    CHECK (neighborhood <> ""),
    CHECK (street <> ""),
    CHECK (number <> ""),
    CHECK (zip_code <> ""),

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

    CHECK (name <> ""),
    CHECK (url <> ""),

    PRIMARY KEY(name)
);

INSERT INTO Portal (name, url) VALUES
("lamudi", "https://lamudi.com.mx"),
("inmuebles24", "https://inmuebles24.com"),
("casasyterrenos", "https://casasyterrenos.com"),
("propiedades", "https://propiedades.com");

DROP TABLE IF EXISTS PublishedProperty;
CREATE TABLE IF NOT EXISTS PublishedProperty (
    property_id INT NOT NULL,

    url     VARCHAR(256) DEFAULT NULL,
    status  ENUM("in_progress", "completed", "failed") DEFAULT "in_progress",
    portal  VARCHAR(64) NOT NULL,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CHECK (url <> ""),

    PRIMARY KEY (portal, property_id),
    FOREIGN KEY (portal) REFERENCES Portal(name),
    FOREIGN KEY (property_id) REFERENCES PortalProp(id)
);

insert into PublishedProperty (property_id, url, status, portal) VALUES
(30, "https://inmuebles24/property/1", "in_progress", "inmuebles24");
        
/* Get the property publications */
SELECT 
    name as portal, 
    ifnull(status, "not_published") as status,
    property_id, PP.url, updated_at, created_at
FROM Portal P
LEFT JOIN PublishedProperty PP
    ON P.name = PP.portal
    AND property_id = 30;
