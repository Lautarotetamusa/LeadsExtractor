package handlers_test

import (
	"leadsextractor/handlers"
	"leadsextractor/models"
	"leadsextractor/store"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
)

func TestUtmDefinitionCRUD(t *testing.T) {
    router := mux.Router{}
    utmHandler.RegisterRoutes(&router)

    noAlphanumeric := `{
        "code": "__TEST__",
        "utm_channel": "ivr"
    }`
    noCode := `{
        "code": "",
        "utm_channel": "ivr"
    }`
    invalidChannel := `{
        "code": "AAAAA89",
        "utm_channel": "invalid-channel"
    }`
    valid := `{
        "code": "CODE",
        "utm_source": "das",
        "utm_medium": "ad",
        "utm_campaing": "dsad",
        "utm_ad": "ewqe",
        "utm_channel": "ivr"
    }`
    utm1 := `{
        "success": true,
        "data": {
            "code": "CODE1",
            "id": 1,
            "utm_source": "aa",
            "utm_medium": "aa",
            "utm_campaign": "aa",
            "utm_ad": "aa",
            "utm_channel": "ivr"
        }
    }`

    utmListRes := handlers.NewDataResponse(utmStore.utms)

    tests := []APITestCase{
        // Get 
        {"get all", "GET", "/utm", "", nil, http.StatusOK, Stringify(utmListRes)},
        {"get not exists", "GET", "/utm/999", "", nil, http.StatusNotFound, "*not found*"},
        {"get one", "GET", "/utm/1", "", nil, http.StatusOK, utm1},
        // Create
        {"only alphanumeric", "POST", "/utm", noAlphanumeric, nil, http.StatusBadRequest, `*only contains alphanumeric characters*`,},
        {"no code", "POST", "/utm", noCode, nil, http.StatusBadRequest, `*code is required*`,},
        {"invalid channel", "POST", "/utm", invalidChannel, nil, http.StatusBadRequest, `*channel must be one of*`,},
        {"create", "POST", "/utm", valid, nil, http.StatusCreated, `*created successfully*`,},
        {"code already exists", "POST", "/utm", valid, nil, http.StatusConflict, `*already exists*`},
        // Update
        {"no pass body", "PUT", "/utm/1", "", nil, http.StatusBadRequest, "",},
        {"invalid json", "PUT", "/utm/1", `{"code": 12345}`, nil, http.StatusBadRequest, "",},
        {"update no exists", "PUT", "/utm/999", `{"code": "CODE2"}`, nil, http.StatusNotFound, "",},
        // Delete
        {"no exists", "DELETE", "/utm/99", "", nil, http.StatusNotFound, "",},
        {"delete", "DELETE", "/utm/2", "", nil, http.StatusOK, "*deleted successfully*",},
    }

    for _, tc := range tests {
        Endpoint(t, &router, tc)
    }
}

type mockUTMStorer struct {
    utms []*models.UtmDefinition
}

func (s *mockUTMStorer) mock() {
    utm := models.UtmDefinition{
        Code: "CODE1",
        Source: models.NullString{String: "aa", Valid: true},
        Medium: models.NullString{String: "aa", Valid: true},
        Campaign: models.NullString{String: "aa", Valid: true},
        Ad: models.NullString{String: "aa", Valid: true},
        Channel: models.NullString{String: "ivr", Valid: true},
    }
    s.Insert(&utm)
}

// GetAll(*[]models.UtmDefinition) error
func (s *mockUTMStorer) GetAll() ([]*models.UtmDefinition, error) {
    return s.utms, nil
}

// GetOne(int) (*models.UtmDefinition, error)
func (s *mockUTMStorer) GetOne(id int) (*models.UtmDefinition, error) {
    for _, utm := range s.utms {
        if utm.Id == id {
            return utm, nil
        }
    }
    return nil, store.NewErr("not found", store.StoreNotFoundErr) 
}

// GetOneByCode(string) (*models.UtmDefinition, error)
func (s *mockUTMStorer) GetOneByCode(code string) (*models.UtmDefinition, error) {
    for _, utm := range s.utms {
        if utm.Code == code {
            return utm, nil
        }
    }
    return nil, store.NewErr("not found", store.StoreNotFoundErr) 
}

func (s *mockUTMStorer) Insert(utm *models.UtmDefinition) (int64, error) {
    if u, _ := s.GetOneByCode(utm.Code); u != nil {
        return 0, store.NewErr("already exists", store.StoreDuplicatedErr)    
    }

    s.utms = append(s.utms, utm)
    id := len(s.utms)
    s.utms[id-1].Id = id
    return int64(id), nil
}

// Update(*models.UtmDefinition) error
func (s *mockUTMStorer) Update(uUTM *models.UtmDefinition) error {
    if u, _ := s.GetOneByCode(uUTM.Code); u != nil {
        return store.NewErr("already exists", store.StoreDuplicatedErr)
    }

    s.utms[uUTM.Id-1] = uUTM
    return nil
}

func (s *mockUTMStorer) Delete(id int) error {
    newList := make([]*models.UtmDefinition, 0)
    for _, utm := range s.utms {
        if utm.Id == id {
            continue
        }
        newList = append(newList, utm)
    }
    if len(s.utms) == len(newList) {
        return store.NewErr("not found", store.StoreNotFoundErr) 
    }

    s.utms = newList
    return nil
}
