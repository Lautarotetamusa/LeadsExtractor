DROP TABLE IF EXISTS PortalProp;
CREATE TABLE IF NOT EXISTS PortalProp(
    id INT NOT NULL AUTO_INCREMENT,

    title VARCHAR(256) NOT NULL,
    price VARCHAR(32) NOT NULL,
    currency CHAR(5) NOT NULL,
    description VARCHAR(512), 
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

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY(id)
);

DROP TABLE IF EXISTS ImageProp;
CREATE TABLE IF NO EXISTS ImageProp(
    property_id INT NOT NULL,

    url VARCHAR(256),

    PRIMARY KEY(property_id),
    FOREIGN KEY(property_id) REFERENCES Property(id)
);
