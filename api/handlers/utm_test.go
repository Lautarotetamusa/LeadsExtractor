package handlers_test

import (
	"leadsextractor/handlers"
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

    utmListRes := handlers.NewDataResponse(utmStore.Utms())

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
