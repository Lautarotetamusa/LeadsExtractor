package handlers_test

import (
	"leadsextractor/handlers"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
)

func TestLeadCRUD(t *testing.T) {
	leadListRes := handlers.NewDataResponse(leadStore.Leads())

	valid := map[string]any{
		"phone":        "5493415854220",
		"name":         "a",
		"asesor_phone": "+5493415854221",
	}
	expected := map[string]any{
		"phone":      "+5493415854220",
		"email":      nil,
		"name":       "a",
		"cotizacion": "",
		"asesor": map[string]any{
			"active": false,
			"email":  "test@gmail.com",
			"name":   "test",
			"phone":  "+5493415854221",
		},
	}

	tests := []APITestCase{
		{"get all", "GET", "/lead", "", nil, http.StatusOK, Stringify(leadListRes)},
		// Create
		{"bad phone number", "POST", "/lead", `{"phone": "12345","name": "a","asesor_phone": "11"}}`, nil, http.StatusBadRequest, `*isn't a valid number*`},
		{"no name", "POST", "/lead", `{"phone": "+5493415854220","asesor_phone": "+5493415854221"}}`, nil, http.StatusBadRequest, `*'Name' failed on the 'required'*`},
		{"create", "POST", "/lead", Stringify(valid), nil, http.StatusCreated, `*Lead creado correctamente*`},
		{"already exists", "POST", "/lead", Stringify(valid), nil, http.StatusConflict, "*already exists*"},
		// Get
		{"get one", "GET", "/lead/5493415854220", "", nil, http.StatusOK, Stringify(handlers.NewDataResponse(expected))},
		{"get one no exists", "GET", "/lead/5493415854222", "", nil, http.StatusNotFound, "*does not exists*"},
		// Update
		{"no pass body", "PUT", "/lead/5493415854220", "", nil, http.StatusBadRequest, ""},
		{"invalid json", "PUT", "/lead/5493415854220", `{"name": 99}`, nil, http.StatusBadRequest, ""},
		{"update no exists", "PUT", "/lead/5493415854222", `{"name": "beatriz"}`, nil, http.StatusNotFound, ""},
	}

	router := mux.Router{}
	leadHandler.RegisterRoutes(&router)
	for _, tc := range tests {
		Endpoint(t, &router, tc)
	}

	upd := map[string]any{
		"name": "juan carlos",
	}
	expected["name"] = "juan carlos"
	updateRes := handlers.NewSuccessResponse(expected, "Lead actualizado correctamente")
	updatedres := handlers.NewDataResponse(expected)
	updateTC := APITestCase{"update", "PUT", "/lead/5493415854220", Stringify(upd), nil, http.StatusCreated, Stringify(updateRes)}
	Endpoint(t, &router, updateTC)
	tc := APITestCase{"get updated", "GET", "/lead/5493415854220", "", nil, http.StatusOK, Stringify(updatedres)}
	Endpoint(t, &router, tc)
}
