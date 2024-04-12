package models

import "database/sql"

type Lead struct {
    Name   string `db:"name"   json:"name"`
    Phone  string `db:"phone"  json:"phone"`
    Email  sql.NullString   `db:"email" json:"email"`
    Asesor Asesor `db:"Asesor" json:"asesor"`
}

type CreateLead struct {
    Name   string `json:"name" validate:"required" db:"name"`
    Phone  string `json:"phone" validate:"required" db:"phone"`
    Email  sql.NullString `json:"email" db:"email"`
    AsesorPhone string `json:"asesor_phone" validate:"required" db:"asesor"`
}

type UpdateLead struct {
    Name   string `db:"name"   json:"name" validate:"required"`
    Email  sql.NullString   `db:"email"  json:"email" validate"required,email"`
}
