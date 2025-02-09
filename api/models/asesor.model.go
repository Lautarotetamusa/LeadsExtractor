package models

import "leadsextractor/pkg/numbers"

type Asesor struct {
    Name   string `db:"name"   json:"name" validate:"required"`
    Phone  numbers.PhoneNumber `db:"phone"  json:"phone" validate:"required"`
    Email  string `db:"email"  json:"email" validate:"required"`
    Active bool   `db:"active" json:"active" validate:"required"`
}

type UpdateAsesor struct {
    Name   *string `db:"name"   json:"name"`
    Phone  *numbers.PhoneNumber `db:"phone"  json:"phone"`
    Email  *string `db:"email"  json:"email"`
    Active *bool  `db:"active" json:"active"`
}
