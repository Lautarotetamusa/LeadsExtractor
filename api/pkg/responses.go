package pkg

import (
	"encoding/json"
	"leadsextractor/store"
	"net/http"
)

type HandlerErrorFn func(w http.ResponseWriter, r *http.Request) error
type HandlerFn func(w http.ResponseWriter, r *http.Request)

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

type ListResponse = struct {
    Success     bool                    `json:"success"`
    Pagination  store.Pagination        `json:"pagination"`
    Data        interface{}             `json:"data"`
}

// used when the response have more than once element
type MultipleError struct {
    Err string `json:"error"`
    Count int `json:"count"`
}

func HandleErrors(fn HandlerErrorFn) HandlerFn {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			ErrorResponse(w, r, err)
		}
	}
}

func ErrorResponse(w http.ResponseWriter, r *http.Request, e error) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(ErrorResponseType{
        Success: false,
        Error: e.Error(),
    })
}

func dataResponse(w http.ResponseWriter, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(DataResponse{
        Success: true,
        Data: data,
    })
}

func successResponse(w http.ResponseWriter, message string, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(SuccessResponse{
        Success: true,
        Message: message,
        Data: data,
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
