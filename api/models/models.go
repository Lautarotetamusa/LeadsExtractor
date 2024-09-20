package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"leadsextractor/numbers"
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

func (ns NullString) Value() (driver.Value, error) {
    if !ns.Valid {
        return nil, nil
    }
    return ns.String, nil
}

func (ns *NullString) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &ns.String)
	ns.Valid = (err == nil)
	return err
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

type NullInt32 sql.NullInt32
func (ni *NullInt32) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ni.Int32)
}

// Scan implements the Scanner interface for NullInt64
func (ni *NullInt32) Scan(value interface{}) error {
	var i sql.NullInt32
	if err := i.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil || value == ""{
		*ni = NullInt32{i.Int32, false}
	} else {
		*ni = NullInt32{i.Int32, true}
	}
	return nil
}

type Source struct {
	Id         int           `db:"id"`
	Tipo       string        `db:"type"`
	PropertyId sql.NullInt16 `db:"property_id"`
}

type Propiedad struct {
    ID        NullInt32  `db:"id" json:"-"` //Lo hacemos null porque en el left join puede no traer
    Portal    string     `db:"portal" json:"-"`
    PortalId  NullString `db:"portal_id" json:"id"`
    Titulo    NullString `json:"titulo" db:"title"`
    Link      NullString `json:"link" db:"url"`
    Precio    NullString `json:"precio" db:"price"`
    Ubicacion NullString `json:"ubicacion" db:"ubication"`
    Tipo      NullString `json:"tipo" db:"tipo"`

    //Campos nuevos agregados para pipedrive (no los guardamos en la DB)
    Bedrooms		string	`json:"bedrooms"`
    Bathrooms		string	`json:"bathrooms"`
    TotalArea		string	`json:"total_area"`
    CoveredArea    string	`json:"covered_area"`
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

type UtmDefinition struct {
    Id          int         `json:"id" db:"id"`
    Code        string      `json:"code" db:"code"`
    Source      NullString   `json:"utm_source" db:"utm_source"`
    Medium      NullString   `json:"utm_medium" db:"utm_medium"`
    Campaign    NullString   `json:"utm_campaign" db:"utm_campaign"`
    Ad          NullString  `json:"utm_ad" db:"utm_ad"`
    Channel     string      `json:"utm_channel" db:"utm_channel"`
}

type Utm struct {
    Source      NullString   `json:"utm_source" db:"utm_source"`
    Medium      NullString   `json:"utm_medium" db:"utm_medium"`
    Campaign    NullString   `json:"utm_campaign" db:"utm_campaign"`
    Ad          NullString  `json:"utm_ad" db:"utm_ad"`
    Channel     string      `json:"utm_channel" db:"utm_channel"`
}

//Este es el objecto que recibimos del python script
type Communication struct {
    Fuente     string    `json:"fuente" db:"fuente"`
    FechaLead  string    `json:"fecha_lead" db:"lead_date"`
    Id         int       `json:"id" db:"id"`
    LeadId     string    `json:"lead_id"` //No guardamos este campo
    Fecha      string    `json:"fecha" db:"created_at"`
    Nombre     string    `json:"nombre" db:"name"`
    Link       string    `json:"link" db:"url"`
    Telefono   numbers.PhoneNumber    `json:"telefono" db:"phone"`
    Email      NullString `json:"email" db:"email"`
    Utm        *Utm      `json:"utm" db:"utm"`
	Cotizacion string    `json:"cotizacion"`
	Asesor     Asesor    `json:"asesor"`
	Propiedad  Propiedad `json:"propiedad"`
	Busquedas  Busquedas `json:"busquedas"`
    IsNew      bool      `json:"is_new" db:"new_lead"`
}
