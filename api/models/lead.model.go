package models

import "leadsextractor/numbers"


type Lead struct {
    Name   string       `db:"name"   json:"name"`
    Phone  numbers.PhoneNumber       `db:"phone"  json:"phone"`
    Email  NullString   `db:"email" json:"email"`
    Asesor Asesor       `db:"Asesor" json:"asesor"`
    Cotizacion string   `db:"cotizacion" json:"cotizacion"`
}

type CreateLead struct {
    Name        string              `json:"name" db:"name" validate:"required"`
    Phone       numbers.PhoneNumber `json:"phone" db:"phone" validate:"required"`
    Email       NullString          `json:"email" db:"email"`
    AsesorPhone numbers.PhoneNumber `json:"asesor_phone" db:"asesor" validate:"required"`
    Cotizacion  string              `db:"cotizacion" json:"cotizacion"`
}

type UpdateLead struct {
    Name   string       `db:"name"   json:"name" validate:"required"`
    Email  NullString   `db:"email"  json:"email" validate:"required,email"`
}
