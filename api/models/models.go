package models

import "database/sql"

type Property struct {
	Id        int    `db:"id"`
	Portal    string `db:"portal"`
	PortalId  string `db:"portal_id"`
	Title     string `db:"title"`
	Price     string `db:"price"`
	Ubication string `db:"ubication"`
	Url       string `db:"url"`
	Tipo      string `db:"tipo"`
}

type Source struct {
	Id         int           `db:"id"`
	Tipo       string        `db:"type"`
	PropertyId sql.NullInt16 `db:"property_id"`
}

type Propiedad struct {
    ID        string `json:"id" db:"id"`
    Titulo    string `json:"titulo" db:"title"`
    Link      string `json:"link" db:"url"`
    Precio    string `json:"precio" db:"price"`
    Ubicacion string `json:"ubicacion" db:"ubication"`
    Tipo      string `json:"tipo" db:"tipo"`
}

type Busquedas struct {
    Zonas            string `json:"zonas" db:"zones"`
    Presupuesto      string `json:"presupuesto"`
    CantidadAnuncios int    `json:"cantidad_anuncios"`
	Contactos        int    `json:"contactos"`
    InicioBusqueda   int    `json:"inicio_busqueda"`
    TotalArea        string `json:"total_area" db:"mt2_builded"`
    CoveredArea      string `json:"covered_area" db:"mt2_terrain"`
    Banios           string `json:"banios" db:"baths"`
    Recamaras        string `json:"recamaras" db:"rooms"`
}

//Este es el objecto que recibimos del python script
type Communication struct {
    Fuente     string    `json:"fuente" db:"fuente"`
    FechaLead  string    `json:"fecha_lead" db:"lead_date"`
    ID         string    `json:"id" db:"id"`
    Fecha      string    `json:"fecha" db:"created_at"`
    Nombre     string    `json:"nombre" db:"name"`
    Link       string    `json:"link" db:"url"`
    Telefono   string    `json:"telefono" db:"phone"`
    Email      string    `json:"email" db:"email"`
	Cotizacion string    `json:"cotizacion"`
	Asesor     Asesor    `json:"asesor"`
	Propiedad  Propiedad `json:"propiedad"`
	Busquedas  Busquedas `json:"busquedas"`
}
