package handlers_test

import (
	"encoding/json"
	"leadsextractor/handlers"
	"leadsextractor/models"
	"leadsextractor/pkg/numbers"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func Stringify(data any) string {
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

func TestAsesorCRUD(t *testing.T) {
	router := mux.Router{}
	asesorHandler.RegisterRoutes(&router)

	asesorList := handlers.NewDataResponse(asesorStore.Asesores())
	leadList := handlers.NewDataResponse(asesorStore.Leads()[asesorStore.Asesores()[0].Phone.String()])

	badPhone := models.Asesor{
		Name:   "juan",
		Phone:  numbers.PhoneNumber("12345"),
		Active: false,
		Email:  "juan@gmail.com",
	}
	noName := models.Asesor{
		Phone:  numbers.PhoneNumber("5493415854220"),
		Active: false,
		Email:  "juan@gmail.com",
	}
	noEmail := models.Asesor{
		Name:   "juan",
		Phone:  numbers.PhoneNumber("5493415854220"),
		Active: false,
	}
	valid := models.Asesor{
		Name:   "juan",
		Phone:  numbers.PhoneNumber("+5493415854220"),
		Active: true,
		Email:  "juan@gmail.com",
	}
	inactive := models.Asesor{
		Name:  "inactive",
		Phone: numbers.PhoneNumber("+5493415555555"),
		Email: "inactive@gmail.com",
	}

	invalidUpd := `{
        "name": 9
    }`

	tests := []APITestCase{
		{"get all", "GET", "/asesor", "", nil, http.StatusOK, Stringify(asesorList)},
		// Create
		{"bad phone", "POST", "/asesor", Stringify(badPhone), nil, http.StatusBadRequest, "*isn't a valid number*"},
		{"no name", "POST", "/asesor", Stringify(noName), nil, http.StatusBadRequest, "*'Name' failed on the 'required'*"},
		{"no email", "POST", "/asesor", Stringify(noEmail), nil, http.StatusBadRequest, "*'Email' failed on the 'required'*"},
		{"inactive", "POST", "/asesor", Stringify(inactive), nil, http.StatusCreated, "*Asesor creado correctamente*"},
		{"create", "POST", "/asesor", Stringify(valid), nil, http.StatusCreated, "*Asesor creado correctamente*"},
		{"already exists", "POST", "/asesor", Stringify(valid), nil, http.StatusConflict, "*already exists*"},
		// Retrieve
		{"get one no exists", "GET", "/asesor/5493415854222", "", nil, http.StatusNotFound, `*does not exists`},
		{"get leads", "GET", "/asesor/5493415554444/leads", "", nil, http.StatusOK, Stringify(leadList)},
		// Update
		{"no pass body", "PUT", "/asesor/+5493415854220", "", nil, http.StatusBadRequest, "*body its required*"},
		{"invalid json", "PUT", "/asesor/+5493415854220", invalidUpd, nil, http.StatusBadRequest, ""},
		{"update no exists", "PUT", "/asesor/5493415854222", "{}", nil, http.StatusNotFound, ""},
	}

	for _, tc := range tests {
		Endpoint(t, &router, tc)
	}

	t.Run("round robin", func(t *testing.T) {
		assert.True(t, rr.Contains(&valid))
		assert.False(t, rr.Contains(&inactive))
	})

	valid.Name = "juan carlos"
	upd := models.UpdateAsesor{
		Name: &valid.Name,
	}
	updateRes := handlers.NewSuccessResponse(valid, "Asesor actualizado correctamente")
	updatedres := handlers.NewDataResponse(valid)
	updateTC := APITestCase{"update", "PUT", "/asesor/+5493415854220", Stringify(upd), nil, http.StatusCreated, Stringify(updateRes)}
	Endpoint(t, &router, updateTC)
	tc := APITestCase{"get updated", "GET", "/asesor/+5493415854220", "", nil, http.StatusOK, Stringify(updatedres)}
	Endpoint(t, &router, tc)
}
