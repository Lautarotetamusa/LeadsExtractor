package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"leadsextractor/bootstrap"
	"leadsextractor/pkg/dependency"

	"github.com/gorilla/mux"
)

type HealthHandler struct {
	app *bootstrap.App
}

func NewHealthHandler(app *bootstrap.App) *HealthHandler {
	return &HealthHandler{app: app}
}

func (h *HealthHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/health", h.Check).Methods(http.MethodGet)
	router.HandleFunc("/health/rebora", h.CheckRebora).Methods(http.MethodGet)
}

// Check corre el HealthCheck de las dependencias propias de este servicio y
// devuelve 200 si todas responden bien, o 503 si alguna falla.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	results := dependency.CheckAll(r.Context(), h.app.Dependencies(), 10*time.Second)
	respondHealth(w, results)
}

// CheckRebora corre el chequeo combinado: este servicio + Portalia
// (desglosado por portal) + Cotizador.
func (h *HealthHandler) CheckRebora(w http.ResponseWriter, r *http.Request) {
	results := h.app.CheckRebora(r.Context())
	respondHealth(w, results)
}

func respondHealth(w http.ResponseWriter, results []dependency.Result) {
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
