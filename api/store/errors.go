package store

import (
	"database/sql"
	"strings"
)

type StoreErrorType int

const (
    StoreNotFoundErr StoreErrorType = iota
    StoreDuplicatedErr
)

type StoreError struct {
    msg string
    Typ StoreErrorType
}

func NewErr(msg string, typ StoreErrorType) StoreError {
    return StoreError{
        msg: msg,
        Typ: typ,
    }
}

func (e StoreError) Error() string {
    return e.msg
}

func SQLNotFound(err error, msg string) error {
    if err == sql.ErrNoRows {
        return StoreError{
            msg: msg,
            Typ: StoreNotFoundErr,
        }
    }
    return err
}

func SQLDuplicated(err error, msg string) error {
    if strings.Contains(err.Error(), "Error 1062") {
        return StoreError{
            msg: msg,
            Typ: StoreDuplicatedErr,
        }
    }
    return err
}
