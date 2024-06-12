package models


type Lead struct {
    Name   string       `db:"name"   json:"name"`
    Phone  string       `db:"phone"  json:"phone"`
    Email  NullString   `db:"email" json:"email"`
    Asesor Asesor       `db:"Asesor" json:"asesor"`
    Cotizacion string   `db:"cotizacion" json:"cotizacion"`
}

type CreateLead struct {
    Name   string       `json:"name" db:"name"`
    Phone  string       `json:"phone" db:"phone"`
    Email  NullString   `json:"email" db:"email"`
    AsesorPhone string  `json:"asesor_phone" db:"asesor"`
    Cotizacion string   `db:"cotizacion" json:"cotizacion"`
}

type UpdateLead struct {
    Name   string       `db:"name"   json:"name" validate:"required"`
    Email  NullString   `db:"email"  json:"email" validate:"required,email"`
}
