package models

import (
	"database/sql"
	"database/sql/driver"
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
	if reflect.TypeOf(value) == nil || value == "" {
		*ni = NullInt32{i.Int32, false}
	} else {
		*ni = NullInt32{i.Int32, true}
	}
	return nil
}

func (ns *NullString) UnmarshalCSV(value string) error {
	if value == "" {
		ns.Valid = false
	} else {
		ns.String = value
		ns.Valid = true
	}
	return nil
}

func (ns NullString) MarshalCSV() (string, error) {
	if ns.Valid {
		return ns.String, nil
	}
	return "", nil
}
