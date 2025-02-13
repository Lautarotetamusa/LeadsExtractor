package handlers_test

import (
	"encoding/json"
	"leadsextractor/handlers"
	"leadsextractor/models"
	"leadsextractor/pkg/numbers"
	"leadsextractor/store"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type mockAsesorStorer struct {
    asesores []*models.Asesor
    leads map[string][]*models.Lead
}

func newMockAsesorStore() *mockAsesorStorer {
    return &mockAsesorStorer{
        asesores: make([]*models.Asesor, 0),
        leads: make(map[string][]*models.Lead),
    }
}

func Stringify(data any) string {
    jsonData, _ := json.Marshal(data)
    return string(jsonData)
}

func TestAsesorCRUD(t *testing.T) {
    router := mux.Router{}
    asesorHandler.RegisterRoutes(&router)

    asesorList := handlers.NewDataResponse(asesorStore.asesores)
    leadList := handlers.NewDataResponse(asesorStore.leads[asesorStore.asesores[0].Phone.String()])

    badPhone := models.Asesor{
        Name: "juan",
        Phone: numbers.PhoneNumber("12345"),
        Active: false,
        Email: "juan@gmail.com",
    }
    noName := models.Asesor{
        Phone: numbers.PhoneNumber("5493415854220"),
        Active: false,
        Email: "juan@gmail.com",
    }
    noEmail := models.Asesor{
        Name: "juan",
        Phone: numbers.PhoneNumber("5493415854220"),
        Active: false,
    }
    valid := models.Asesor{
        Name: "juan",
        Phone: numbers.PhoneNumber("+5493415854220"),
        Active: true,
        Email: "juan@gmail.com",
    }
    inactive := models.Asesor{
        Name: "inactive",
        Phone: numbers.PhoneNumber("+5493415555555"),
        Email: "inactive@gmail.com",
    }

    invalidUpd := `{
        "name": 9
    }`

    tests := []APITestCase{
        {"get all", "GET", "/asesor", "", nil, http.StatusOK, Stringify(asesorList)},
        // Create
        {"bad phone", "POST", "/asesor", Stringify(badPhone), nil, http.StatusBadRequest, "*isn't a valid number*",},
        {"no name", "POST", "/asesor", Stringify(noName), nil, http.StatusBadRequest, "*'Name' failed on the 'required'*",},
        {"no email", "POST", "/asesor", Stringify(noEmail), nil, http.StatusBadRequest, "*'Email' failed on the 'required'*",},
        {"inactive", "POST", "/asesor", Stringify(inactive), nil, http.StatusCreated, "*Asesor creado correctamente*",},
        {"create", "POST", "/asesor", Stringify(valid), nil, http.StatusCreated, "*Asesor creado correctamente*",},
        {"already exists", "POST", "/asesor", Stringify(valid), nil, http.StatusConflict, "*already exists*",},
        // Retrieve
        {"get one no exists", "GET", "/asesor/5493415854222", "", nil, http.StatusNotFound, `*does not exists`,},
        {"get leads", "GET", "/asesor/5493415554444/leads", "", nil, http.StatusOK, Stringify(leadList),},
        // Update
        {"no pass body", "PUT", "/asesor/+5493415854220", "", nil, http.StatusBadRequest, "*body its required*",},
        {"invalid json", "PUT", "/asesor/+5493415854220", invalidUpd, nil, http.StatusBadRequest, "",},
        {"update no exists", "PUT", "/asesor/5493415854222", "{}", nil, http.StatusNotFound, "",},
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
    updateTC := APITestCase{"update", "PUT", "/asesor/+5493415854220", Stringify(upd), nil, http.StatusCreated, Stringify(updateRes),}
    Endpoint(t, &router, updateTC)
    tc := APITestCase{"get updated", "GET", "/asesor/+5493415854220", "", nil, http.StatusOK, Stringify(updatedres),}
    Endpoint(t, &router, tc)
}

func (s *mockAsesorStorer) mock() {
    // Name   string `db:"name"   json:"name" validate:"required"`
    // Phone  numbers.PhoneNumber `db:"phone"  json:"phone" validate:"required"`
    // Email  string `db:"email"  json:"email" validate:"required"`
    // Active bool   `db:"active" json:"active" validate:"required"`
    s.asesores = []*models.Asesor{
        {"juan", numbers.PhoneNumber("5493415554444"), "juan@gmail.com", true},
        {"carlos", numbers.PhoneNumber("5493415556666"), "carlos@gmail.com", true},
        {"jose", numbers.PhoneNumber("5493415557777"), "jose@gmail.com", true},
        {"maria", numbers.PhoneNumber("5493415558888"), "maria@gmail.com", false},
        {"juana", numbers.PhoneNumber("5493415559999"), "juana@gmail.com", true},
    }

    email := models.NullString{String: "lead@gmail.com", Valid: true} 
    s.leads["5493415554444"] = []*models.Lead{
        {"lead1", numbers.PhoneNumber("5493414449999"), email, *s.asesores[0], ""},
        {"lead2", numbers.PhoneNumber("5493414448888"), email, *s.asesores[0], ""},
        {"lead3", numbers.PhoneNumber("5493414447777"), email, *s.asesores[0], ""},
    }
    s.leads["5493415559999"] = []*models.Lead{
        {"lead1", numbers.PhoneNumber("5493414449999"), email, *s.asesores[4], ""},
        {"lead2", numbers.PhoneNumber("5493414448888"), email, *s.asesores[4], ""},
        {"lead3", numbers.PhoneNumber("5493414447777"), email, *s.asesores[4], ""},
    }
}

func (s *mockAsesorStorer) GetAll() ([]*models.Asesor, error) {
    return s.asesores, nil
}

func (s *mockAsesorStorer) GetAllActive() ([]*models.Asesor, error) {
    return s.asesores, nil
}

func (s *mockAsesorStorer) GetLeads(phone string) ([]*models.Lead, error) {
    a, _ := s.GetOne(phone)
    leads := s.leads[a.Phone.String()]
    return leads, nil
}

func (s *mockAsesorStorer) GetOne(phone string) (*models.Asesor, error) {
    for _, a := range s.asesores {
        if a.Phone.String() == phone {
            return a, nil
        }
    }
    return nil, store.NewErr("does not exists", store.StoreNotFoundErr) 
}

func (s *mockAsesorStorer) GetFromEmail(email string) (*models.Asesor, error) {
    for _, a := range s.asesores {
        if a.Email == email {
            return a, nil
        }
    }
    return nil, store.NewErr("not found", store.StoreNotFoundErr) 
}

func (s *mockAsesorStorer) Insert(asesor *models.Asesor) error {
    if l, _ := s.GetOne(asesor.Phone.String()); l != nil {
        return store.NewErr("already exists", store.StoreDuplicatedErr)
    }

    s.asesores = append(s.asesores, asesor)
    return nil
}

func (s *mockAsesorStorer) Update(asesor *models.Asesor) error {
    for i, a := range s.asesores {
        if a.Phone.String() == asesor.Phone.String() {
            s.asesores[i] = asesor
            return nil
        }
    }
    return store.NewErr("not found", store.StoreNotFoundErr)
}

func (s *mockAsesorStorer) Delete(asesor *models.Asesor) error {
    return nil
}
