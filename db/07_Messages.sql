CREATE TABLE IF NOT EXISTS Message (
    id INT NOT NULL AUTO_INCREMENT,
    id_communication INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    text TEXT NOT NULL,
    wamid VARCHAR(64) DEFAULT NULL,

    CHECK (text != ""),
    
    PRIMARY KEY (id),
    FOREIGN KEY (id_communication) REFERENCES Communication(id)
);
