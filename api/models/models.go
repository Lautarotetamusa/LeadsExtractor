package models

import (
	"database/sql"
	"encoding/json"
	"reflect"
)

type NullString sql.NullString
func (s *NullString) MarshalJSON() ([]byte, error) {
    if !s.Valid {
        return []byte("null"), nil
    }
    return json.Marshal(s.String)
}

func (ns *NullString) Scan(value interface{}) error {
	var s sql.NullString
	if err := s.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*ns = NullString{s.String, false}
	} else {
		*ns = NullString{s.String, true}
	}

	return nil
}

type NullInt16 sql.NullInt16
func (ni *NullInt16) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ni.Int16)
}

// Scan implements the Scanner interface for NullInt64
func (ni *NullInt16) Scan(value interface{}) error {
	var i sql.NullInt16
	if err := i.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil {
		*ni = NullInt16{i.Int16, false}
	} else {
		*ni = NullInt16{i.Int16, true}
	}
	return nil
}

type Property struct {
    Id        NullInt16  `db:"id" json:"-"`
    Portal    NullString `db:"portal" json:"-"`
    PortalId  NullString `db:"portal_id" json:"id"`
    Title     NullString `db:"title" json:"title"`
    Price     NullString `db:"price" json:"price"`
    Ubication NullString `db:"ubication" json:"ubication"`
    Url       NullString `db:"url" json:"url"`
    Tipo      NullString `db:"tipo" json:"tipo"`
}

type Source struct {
	Id         int           `db:"id"`
	Tipo       string        `db:"type"`
	PropertyId sql.NullInt16 `db:"property_id"`
}

type Propiedad struct {
    ID        string `json:"portal_id" db:"id"`
    Titulo    string `json:"titulo" db:"title"`
    Link      string `json:"link" db:"url"`
    Precio    string `json:"precio" db:"price"`
    Ubicacion string `json:"ubicacion" db:"ubication"`
    Tipo      string `json:"tipo" db:"tipo"`
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

type CommunicationDB struct {
    Fuente     string    `json:"fuente" db:"fuente"`
    FechaLead  string    `json:"fecha_lead" db:"lead_date"`
    ID         string    `json:"id" db:"id"`
    Fecha      string    `json:"fecha" db:"created_at"`
    Nombre     string    `json:"nombre" db:"name"`
    Link       string    `json:"link" db:"url"`
    Telefono   string    `json:"telefono" db:"phone"`
    Email      NullString `json:"email" db:"email"`
	Cotizacion string    `json:"cotizacion"`
	Asesor     Asesor    `json:"asesor"`
	Propiedad  Property  `json:"propiedad"`
	Busquedas  Busquedas `json:"busquedas"`
    IsNew       bool     `json:"is_new" db:"new_lead"`
}
