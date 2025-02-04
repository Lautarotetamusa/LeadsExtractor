package handlers_test

import (
	"database/sql"
	"errors"
	"leadsextractor/models"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
)

func TestUtmDefinitionCRUD(t *testing.T) {
    tests := []APITestCase{
        {"bad phone number", "POST", "/utm", `{"phone": "12345","name": "a","asesor_phone": "11"}}`, nil, http.StatusBadRequest, `*isn't a valid number*`,},
        {"no name", "POST", "/utm", `{"phone": "+5493415854220","asesor_phone": "+5493415854221"}}`, nil, http.StatusBadRequest, `*no tiene nombre*`,},
        {"create", "POST", "/utm", `{"phone": "+5493415854220","name": "a","asesor_phone": "+5493415854221"}}`, nil, http.StatusCreated, `*UtmDefinition creado correctamente*`,},
        {"get one no exists", "GET", "/utm/5493415854222", "", nil, http.StatusNotFound, `*no rows in result set`,}, }

    router := mux.Router{}
    utmHandler.RegisterRoutes(&router)
    for _, tc := range tests {
        Endpoint(t, &router, tc)
    }
}

type mockUTMStorer struct {
    utms []models.UtmDefinition
}

// GetAll(*[]models.UtmDefinition) error
func (s mockUTMStorer) GetAll(utms *[]models.UtmDefinition) error {
    utms = &s.utms
    return nil
}

// GetOne(int) (*models.UtmDefinition, error)
func (s mockUTMStorer) GetOne(id int) (*models.UtmDefinition, error) {
    for _, utm := range s.utms {
        if utm.Id == id {
            return &utm, nil
        }
    }
    return nil, sql.ErrNoRows
}

// GetOneByCode(string) (*models.UtmDefinition, error)
func (s mockUTMStorer) GetOneByCode(code string) (*models.UtmDefinition, error) {
    for _, utm := range s.utms {
        if utm.Code == code {
            return &utm, nil
        }
    }
    return nil, sql.ErrNoRows
}

func (s mockUTMStorer) Insert(utm *models.UtmDefinition) (int64, error) {
    s.utms = append(s.utms, *utm) 
    return 101, nil
}

// Update(*models.UtmDefinition) error
func (s mockUTMStorer) Update(uUTM *models.UtmDefinition) error {
    for i, utm := range s.utms {
        if utm.Id == uUTM.Id {
            s.utms[i] = *uUTM
        }
    }
    return errors.New("not found")
}
