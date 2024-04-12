CREATE USER IF NOT EXISTS 'teti'@'%' IDENTIFIED BY 'Lautaro123.';
GRANT ALL PRIVILEGES ON *.* TO 'teti'@'%' WITH GRANT OPTION;

CREATE DATABASE IF NOT EXISTS LeadsExtractor;
USE LeadsExtractor;

CREATE TABLE Asesor(
    name  VARCHAR(64) NOT NULL,
    phone CHAR(16) NOT NULL,
    active BOOLEAN,

    PRIMARY KEY (phone)
);

CREATE TABLE Leads(
    name  VARCHAR(64) NOT NULL,
    phone CHAR(16) NOT NULL,
    email VARCHAR(64) DEFAULT NULL,
    asesor CHAR(16) NOT NULL,

    PRIMARY KEY (phone),
    FOREIGN KEY (asesor) REFERENCES Asesor(phone)
);
