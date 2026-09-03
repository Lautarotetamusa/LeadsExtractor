package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"leadsextractor/pkg/dependency"

	"github.com/gorilla/mux"
)

type HealthHandler struct {
	deps []dependency.Dependency
}

func NewHealthHandler(deps []dependency.Dependency) *HealthHandler {
	return &HealthHandler{deps: deps}
}

func (h *HealthHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/health", h.Check).Methods(http.MethodGet)
}

// Check corre el HealthCheck de todas las dependencias externas en paralelo
// y devuelve 200 si todas responden bien, o 503 si alguna falla.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	results := dependency.CheckAll(r.Context(), h.deps, 10*time.Second)

	status := http.StatusOK
	if !dependency.AllOK(results) {
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(DataResponse{
		Success: status == http.StatusOK,
		Data:    results,
	})
}
