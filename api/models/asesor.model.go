package models

type Asesor struct {
    Name   string `db:"name"   json:"name" validate:"required"`
    Phone  string `db:"phone"  json:"phone" validate:"required"`
    Active bool   `db:"active" json:"active" validate"required"`
}

type UpdateAsesor struct {
    Name   string `db:"name"   json:"name" validate:"required"`
    Active bool   `db:"active" json:"active" validate"required"`
}
