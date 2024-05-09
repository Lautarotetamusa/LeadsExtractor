package models

import "database/sql"

/*
CREATE TABLE Property(
    id INT NOT NULL AUTO_INCREMENT,
    portal ENUM("inmuebles24", "lamudi", "casasyterrenos", "propiedades") NOT NULL,

    portal_id VARCHAR(128) DEFAULT NULL,
    title VARCHAR(256) DEFAULT NULL,
    price VARCHAR(32) DEFAULT NULL,
    ubication VARCHAR(256) DEFAULT NULL,
    url VARCHAR(256) DEFAULT NULL,
    tipo VARCHAR(32) DEFAULT NULL,
    zones VARCHAR(256) DEFAULT NULL,
    mt2_terrain VARCHAR(32) DEFAULT NULL,
    mt2_builded VARCHAR(32) DEFAULT NULL,
    baths VARCHAR(32) DEFAULT NULL,
    rooms VARCHAR(32) DEFAULT NUll,

    PRIMARY KEY (id)
);*/
type Property struct {
    Id          int    `db:"id"`
    Portal      string `db:"portal"`
    PortalId    string `db:"portal_id"`
    Title       string `db:"title"`
    Price       string `db:"price"`
    Ubication   string `db:"ubication"`
    Url         string `db:"url"`
    Tipo		string `db:"tipo"`
}

/*
CREATE TABLE Source(
    id INT NOT NULL AUTO_INCREMENT,
    type ENUM("whatsapp", "ivr", "property") NOT NULL, 
    property_id INT DEFAULT NULL,
    
    PRIMARY KEY (id),
    FOREIGN KEY (property_id) REFERENCES Property(id)
);
*/
type Source struct{
    Id          int             `db:"id"`
    Tipo        string          `db:"type"`
    PropertyId  sql.NullInt16   `db:"property_id"`
}

//Este es el objecto que recibimos del python script
type Communication struct {
    Fuente      string `json:"fuente"`
    FechaLead   string `json:"fecha_lead"`
    ID          string `json:"id"`
    Fecha       string `json:"fecha"`
    Nombre      string `json:"nombre"`
    Link        string `json:"link"`
    Telefono    string `json:"telefono"`
    Email       string `json:"email"`
    Cotizacion  string `json:"cotizacion"`
    Asesor      Asesor `json:"asesor"`
    Propiedad   struct {
        ID          string `json:"id"`
        Titulo      string `json:"titulo"`
        Link        string `json:"link"`
        Precio      string `json:"precio"`
        Ubicacion   string `json:"ubicacion"`
        Tipo        string `json:"tipo"`
    } `json:"propiedad"`
    Busquedas struct {
        Zonas            string `json:"zonas"`
        Presupuesto      string `json:"presupuesto"`
        CantidadAnuncios int    `json:"cantidad_anuncios"`
        Contactos        int    `json:"contactos"`
        InicioBusqueda   int    `json:"inicio_busqueda"`
        TotalArea        string `json:"total_area"`
        CoveredArea      string `json:"covered_area"`
        Banios           string `json:"banios"`
        Recamaras        string `json:"recamaras"`
    } `json:"busquedas"`
}
