package models

import (
	"database/sql"
	"leadsextractor/numbers"
)

type Source struct {
	Id         int           `db:"id"`
	Tipo       string        `db:"type"`
	PropertyId sql.NullInt16 `db:"property_id"`
}

type Propiedad struct {
    ID        NullInt32  `json:"-"          db:"id" `       // Lo hacemos null porque en el left join puede no traer
    Portal    string     `json:"-"          db:"portal"     csv:"fuente"`
    PortalId  NullString `json:"id"         db:"portal_id"  csv:"id"`
    Titulo    NullString `json:"titulo"     db:"title"      csv:"titulo"`
    Link      NullString `json:"link"       db:"url"        csv:"url"`
    Precio    NullString `json:"precio"     db:"price"      csv:"precio"`
    Ubicacion NullString `json:"ubicacion"  db:"ubication"  csv:"ubicacion"`
    Tipo      NullString `json:"tipo"       db:"tipo"`

    // Campos nuevos agregados para pipedrive (no los guardamos en la DB)
    Bedrooms		string	`json:"bedrooms" csv:"habitaciones"`
    Bathrooms		string	`json:"bathrooms" csv:"banios"`
    TotalArea		string	`json:"total_area" csv:"area"`
    CoveredArea     string	`json:"covered_area"`
}

type Busquedas struct {
    Zonas            NullString `json:"zonas" db:"zones"`
    Presupuesto      string `json:"presupuesto"`
    CantidadAnuncios int    `json:"cantidad_anuncios"`
	Contactos        int    `json:"contactos"`
    InicioBusqueda   int    `json:"inicio_busqueda"`
    TotalArea        NullString `json:"total_area" db:"mt2_builded"`
    CoveredArea      NullString `json:"covered_area" db:"mt2_terrain"`
    Banios           NullString `json:"banios" db:"baths"`
    Recamaras        NullString `json:"recamaras" db:"rooms"`
}

type Communication struct {
    Fuente     string    `json:"fuente" db:"fuente" csv:"fuente"`
    FechaLead  string    `json:"fecha_lead" db:"lead_date" csv:"fecha"`
    Id         int       `json:"id" db:"id"`
    LeadId     string    `json:"lead_id"`
    Fecha      string    `json:"fecha" db:"created_at"`
    Nombre     string    `json:"nombre" db:"name" csv:"nombre"`
    Link       string    `json:"link" db:"url"`
    Telefono   numbers.PhoneNumber    `json:"telefono" db:"phone" csv:"telefono"`
    Email      NullString `json:"email" db:"email" csv:"email"`
    Utm        Utm      `json:"utm" db:"utm"`
    Cotizacion string    `json:"cotizacion"`
    Asesor     Asesor    `json:"asesor"`
    Propiedad  Propiedad `json:"propiedad" csv:"propiedad"`
    Busquedas  Busquedas `json:"busquedas"`
    IsNew      bool      `json:"is_new" db:"new_lead"`
    Message    NullString `json:"message" csv:"mensaje"`
    Wamid      NullString 
}
