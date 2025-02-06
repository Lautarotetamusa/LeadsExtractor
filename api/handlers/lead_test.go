package handlers_test

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"leadsextractor/models"
	"leadsextractor/numbers"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestLeadCRUD(t *testing.T) {
    tests := []APITestCase{
        {"bad phone number", "POST", "/lead", `{"phone": "12345","name": "a","asesor_phone": "11"}}`, nil, http.StatusBadRequest, `*isn't a valid number*`,},
        {"no name", "POST", "/lead", `{"phone": "+5493415854220","asesor_phone": "+5493415854221"}}`, nil, http.StatusBadRequest, `*'Name' failed on the 'required'*`,},
        {"create", "POST", "/lead", `{"phone": "+5493415854220","name": "a","asesor_phone": "+5493415854221"}}`, nil, http.StatusCreated, `*Lead creado correctamente*`,},
        {"get one no exists", "GET", "/lead/5493415854222", "", nil, http.StatusNotFound, `*does not exists`,},
    }

    router := mux.Router{}
    leadHandler.RegisterRoutes(&router)
    for _, tc := range tests {
        Endpoint(t, &router, tc)
    }
}

type mockLeadStorer struct {
    leads []models.Lead
}

func (s mockLeadStorer) GetAll() (*[]models.Lead, error) {
    return &s.leads, nil
}

func (s mockLeadStorer) GetOne(phone numbers.PhoneNumber) (*models.Lead, error) {
    for _, lead := range s.leads {
        if lead.Phone == phone {
            return &lead, nil
        }
    }
    return nil, sql.ErrNoRows
}

func (s mockLeadStorer) Insert(createLead *models.CreateLead) (*models.Lead, error) {
    lead := models.Lead{
        Name: createLead.Name,
        Phone: createLead.Phone,
        Email: createLead.Email,
        Cotizacion: createLead.Cotizacion,
        Asesor: models.Asesor{
            Name: "test",
            Phone: createLead.AsesorPhone,
        },
    }

    s.leads = append(s.leads, lead) 
    return &lead, nil
}

func (s mockLeadStorer) Update(uLead *models.Lead, phone numbers.PhoneNumber) error {
    for i, lead := range s.leads {
        if lead.Phone == phone {
            s.leads[i] = *uLead
        }
    }
    return errors.New("not found")
}

func (s mockLeadStorer) UpdateAsesor(phone numbers.PhoneNumber, a *models.Asesor) error {
    for i, lead := range s.leads {
        if lead.Phone == phone {
            s.leads[i].Asesor = *a
        }
    }
    return errors.New("not found")
}

func Endpoint(t *testing.T, router *mux.Router, tc APITestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		req, _ := http.NewRequest(tc.Method, tc.URL, bytes.NewBufferString(tc.Body))
		if tc.Header != nil {
			req.Header = tc.Header
		}
		w := httptest.NewRecorder()
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}

        router.ServeHTTP(w, req)
        res := w.Result()
        defer res.Body.Close()
        response, _ := io.ReadAll(res.Body)

		assert.Equal(t, tc.WantStatus, res.StatusCode, "status mismatch")
		if tc.WantResponse != "" {
			pattern := strings.Trim(tc.WantResponse, "*")

			if pattern != tc.WantResponse {
				assert.Contains(t, string(response), pattern, "response mismatch")
			} else {
				assert.JSONEq(t, tc.WantResponse, string(response), "mismatch")
			}
		}
	})
}
