package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"leadsextractor/store"
	"log/slog"
	"net/http"
)

type HandlerErrorFn func(w http.ResponseWriter, r *http.Request) error
type HandlerFn func(w http.ResponseWriter, r *http.Request)

type APIError struct {
    Status  int
    Msg     string
}

func (e APIError) Error() string {
    return fmt.Sprintf("%d - %s", e.Status, e.Msg)
}

func ErrNotFound(msg string) APIError{
    return APIError{
        Status: http.StatusNotFound,
        Msg: msg,
    }
}

func ErrBadRequest(msg string) APIError{
    return APIError{
        Status: http.StatusBadRequest,
        Msg: msg,
    }
}

func ErrDuplicated(msg string) APIError{
    return APIError{
        Status: http.StatusConflict,
        Msg: msg,
    }
}

var ErrInternal = APIError{
    Status: http.StatusInternalServerError,
    Msg:    "internal server error",
}

type ErrorResponseType struct {
    Success bool    `json:"success"`
    Error   string  `json:"error"`
}

type DataResponse struct {
    Success     bool        `json:"success"`
    Data        interface{} `json:"data"` 
}

type SuccessResponse struct {
    Success     bool        `json:"success"`
    Message     string      `json:"message"`
    Data        interface{} `json:"data"` 
}

type ListResponse struct {
    Success     bool                    `json:"success"`
    Pagination  store.Pagination        `json:"pagination"`
    Data        interface{}             `json:"data"`
}

// used when the response have more than once element
type MultipleError struct {
    Err string `json:"error"`
    Count int `json:"count"`
}

func NewSuccessResponse(data interface{}, msg string) *SuccessResponse {
    return &SuccessResponse{
        Message: msg,
        Success: true, 
        Data: data,
    }
}

func NewDataResponse(data interface{}) *DataResponse {
    return &DataResponse{
        Success: true, 
        Data: data,
    }
}

func HandleErrors(fn HandlerErrorFn) HandlerFn {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
            slog.Error(fmt.Sprintf("%#v\n", err))

            apiErr, isApiErr := err.(APIError)
            if !isApiErr {
                if storeErr, isStoreErr := err.(store.StoreError); isStoreErr {
                    apiErr = store2APIErr(storeErr) 
                }else{
                    apiErr = ErrInternal
                }
            }

            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(apiErr.Status)

            json.NewEncoder(w).Encode(ErrorResponseType{
                Success: false,
                Error: apiErr.Error(),
            })
		}
	}
}

func dataResponse(w http.ResponseWriter, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(DataResponse{
        Success: true,
        Data: data,
    })
}

func createdResponse(w http.ResponseWriter, message string, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)

    json.NewEncoder(w).Encode(SuccessResponse{
        Success: true,
        Message: message,
        Data: data,
    })
}

func messageResponse(w http.ResponseWriter, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    json.NewEncoder(w).Encode(SuccessResponse{
        Success: true,
        Message: message,
    })
}

// Collect all the key-values (errors) into a []MultipleError
func collectMultipleErrors(errorSet map[string]int) []MultipleError {
    var errors []MultipleError
    for k, v := range errorSet {
        errors = append(errors, MultipleError{
            Err: k,
            Count: v,
        })
    }
    return errors
}

// Converts the StoreError type to APIError
func store2APIErr(s store.StoreError) APIError {
    switch s.Typ {
    case store.StoreNotFoundErr:
        return ErrNotFound(s.Error())
    case store.StoreDuplicatedErr:
        return ErrDuplicated(s.Error())
    }
    return ErrInternal
}

func jsonErr(e error) APIError {
    if e == io.EOF {
        return ErrBadRequest("body its required")
    }
    return ErrBadRequest(e.Error())
}
