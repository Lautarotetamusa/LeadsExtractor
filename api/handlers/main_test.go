package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"leadsextractor/handlers"
	"leadsextractor/store"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type RouterFunc func (w http.ResponseWriter, r *http.Request) error

type APITestCase struct {
	Name         string
	Method, URL  string
	Body         string
	Header       http.Header
	WantStatus   int
	WantResponse string
}

var (
    utmHandler  *handlers.UTMHandler
    leadHandler *handlers.LeadHandler
    commHandler *handlers.CommunicationHandler
    asesorHandler *handlers.AsesorHandler

    leadStore   mockLeadStorer
    utmStore    mockUTMStorer
    asesorStore mockAsesorStorer
)

func TestMain(t *testing.M) {
    leadStore = mockLeadStorer{}
    utmStore = mockUTMStorer{}
    asesorStore = *newMockAsesorStore()
    leadStore.mock()
    utmStore.mock()
    asesorStore.mock()

    asesores, _ := asesorStore.GetAll()
    rr := store.NewRoundRobin(*asesores)

    leadService := handlers.NewLeadService(&leadStore)
    asesorService := handlers.NewAsesorService(&asesorStore, &leadStore, rr)

    leadHandler = handlers.NewLeadHandler(leadService) 
    utmHandler = handlers.NewUTMHandler(&utmStore)
    asesorHandler = handlers.NewAsesorHandler(asesorService)

    t.Run()
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

        if tc.WantStatus != res.StatusCode {
            println(req.URL.String(), string(response))
        }
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

func GetResponse(router *mux.Router, tc APITestCase, data any) {
    req, _ := http.NewRequest(tc.Method, tc.URL, bytes.NewBufferString(tc.Body))
    if tc.Header != nil {
        req.Header = tc.Header
    }
    w := httptest.NewRecorder()
    req.Header.Set("Content-Type", "application/json")

    router.ServeHTTP(w, req)
    res := w.Result()
    defer res.Body.Close()
    json.NewDecoder(res.Body).Decode(data)
}
