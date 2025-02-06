package store

import (
	"database/sql"
	"strings"
)

type ErrStore struct {
    msg string
}

func NewErr(msg string) ErrStore {
    return ErrStore{
        msg: msg,
    }
}

func (e ErrStore) Error() string {
    return e.msg
}

func SQLNotFound(err error, msg string) error {
    if err == sql.ErrNoRows {
        return NewErr(msg) 
    }
    return err
}

func SQLDuplicated(err error, msg string) error {
    if strings.Contains(err.Error(), "Error 1062") {
        return NewErr(msg)
    }
    return err
}
