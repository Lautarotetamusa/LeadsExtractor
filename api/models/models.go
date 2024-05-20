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
	ID        string `json:"id"`
	Titulo    string `json:"titulo"`
	Link      string `json:"link"`
	Precio    string `json:"precio"`
	Ubicacion string `json:"ubicacion"`
	Tipo      string `json:"tipo"`
}

type Busquedas struct {
	Zonas            string `json:"zonas"`
	Presupuesto      string `json:"presupuesto"`
	CantidadAnuncios int    `json:"cantidad_anuncios"`
	Contactos        int    `json:"contactos"`
	InicioBusqueda   int    `json:"inicio_busqueda"`
	TotalArea        string `json:"total_area"`
	CoveredArea      string `json:"covered_area"`
	Banios           string `json:"banios"`
	Recamaras        string `json:"recamaras"`
}

//Este es el objecto que recibimos del python script
type Communication struct {
	Fuente     string    `json:"fuente"`
	FechaLead  string    `json:"fecha_lead"`
	ID         string    `json:"id"`
	Fecha      string    `json:"fecha"`
	Nombre     string    `json:"nombre"`
	Link       string    `json:"link"`
	Telefono   string    `json:"telefono"`
	Email      string    `json:"email"`
	Cotizacion string    `json:"cotizacion"`
	Asesor     Asesor    `json:"asesor"`
	Propiedad  Propiedad `json:"propiedad"`
	Busquedas  Busquedas `json:"busquedas"`
}
