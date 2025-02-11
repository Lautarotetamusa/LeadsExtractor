package handlers_test

import (
	"leadsextractor/handlers"
	"leadsextractor/models"
	"leadsextractor/pkg/numbers"
	"leadsextractor/store"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
)

func TestLeadCRUD(t *testing.T) {
    leadListRes := handlers.NewDataResponse(leadStore.leads)

    valid := map[string]any{
        "phone": "5493415854220",
        "name": "a",
        "asesor_phone": "+5493415854221",
    }
    expected := map[string]any{
        "phone": "+5493415854220",
        "email": nil,
        "name": "a",
        "cotizacion": "",
        "asesor": map[string]any{
            "active": false,
            "email": "test@gmail.com",
            "name": "test",
            "phone": "+5493415854221",
        },
    }

    tests := []APITestCase{
        {"get all", "GET", "/lead", "", nil, http.StatusOK, Stringify(leadListRes)},
        // Create
        {"bad phone number", "POST", "/lead", `{"phone": "12345","name": "a","asesor_phone": "11"}}`, nil, http.StatusBadRequest, `*isn't a valid number*`,},
        {"no name", "POST", "/lead", `{"phone": "+5493415854220","asesor_phone": "+5493415854221"}}`, nil, http.StatusBadRequest, `*'Name' failed on the 'required'*`,},
        {"create", "POST", "/lead", Stringify(valid), nil, http.StatusCreated, `*Lead creado correctamente*`,},
        // Get
        {"get one", "GET", "/lead/5493415854220", "", nil, http.StatusOK, Stringify(handlers.NewDataResponse(expected)),},
        {"get one no exists", "GET", "/lead/5493415854222", "", nil, http.StatusNotFound, `*does not exists`,},
        // Update
        {"no pass body", "PUT", "/lead/5493415854220", "", nil, http.StatusBadRequest, "",},
        {"invalid json", "PUT", "/lead/5493415854220", `{"name": 99}`, nil, http.StatusBadRequest, "",},
        {"update no exists", "PUT", "/lead/5493415854222", `{"name": "beatriz"}`, nil, http.StatusNotFound, "",},
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
    updateTC := APITestCase{"update", "PUT", "/lead/5493415854220", Stringify(upd), nil, http.StatusCreated, Stringify(updateRes),}
    Endpoint(t, &router, updateTC)
    tc := APITestCase{"get updated", "GET", "/lead/5493415854220", "", nil, http.StatusOK, Stringify(updatedres),}
    Endpoint(t, &router, tc)
}

type mockLeadStorer struct {
    leads []models.Lead
}

func (s *mockLeadStorer) mock() {
    s.Insert(&models.CreateLead{
        Name: "test",
        Phone: numbers.PhoneNumber("5493415854220"),
        Email: models.NullString{String: "test@gmail.com", Valid: true},
        AsesorPhone: numbers.PhoneNumber("5493415854212"),
        Cotizacion: "",
    })
}

func (s *mockLeadStorer) GetAll() (*[]models.Lead, error) {
    return &s.leads, nil
}

func (s *mockLeadStorer) GetOne(phone numbers.PhoneNumber) (*models.Lead, error) {
    for _, lead := range s.leads {
        if lead.Phone == phone {
            return &lead, nil
        }
    }
    return nil, store.NewErr("does not exists", store.StoreNotFoundErr)
}

func (s *mockLeadStorer) Insert(createLead *models.CreateLead) (*models.Lead, error) {
    if l, _ := s.GetOne(createLead.Phone); l != nil {
        return nil, store.NewErr("already exists", store.StoreDuplicatedErr)
    }

    lead := models.Lead{
        Name: createLead.Name,
        Phone: createLead.Phone,
        Email: createLead.Email,
        Cotizacion: createLead.Cotizacion,
        Asesor: models.Asesor{
            Name: "test",
            Phone: createLead.AsesorPhone,
            Email: "test@gmail.com", // Required field
        },
    }

    s.leads = append(s.leads, lead) 
    return &lead, nil
}

func (s *mockLeadStorer) Update(uLead *models.Lead) error {
    for i, lead := range s.leads {
        if lead.Phone.String() == uLead.Phone.String() {
            s.leads[i] = *uLead
            return nil
        }
    }
    return store.NewErr("not found", store.StoreNotFoundErr)
}

func (s *mockLeadStorer) UpdateAsesor(phone numbers.PhoneNumber, a *models.Asesor) error {
    for i, lead := range s.leads {
        if lead.Phone == phone {
            s.leads[i].Asesor = *a
        }
    }
    return store.NewErr("not found", store.StoreNotFoundErr)
}
