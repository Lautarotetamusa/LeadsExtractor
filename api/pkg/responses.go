package pkg

import (
	"encoding/json"
	"net/http"
)

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
