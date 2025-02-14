package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"leadsextractor/handlers"
	"leadsextractor/mocks"
	"leadsextractor/models"
	"leadsextractor/pkg/roundrobin"
	"log/slog"
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
    utmHandler      *handlers.UTMHandler
    leadHandler     *handlers.LeadHandler
    commHandler     *handlers.CommunicationHandler
    asesorHandler   *handlers.AsesorHandler

    leadStore   mocks.MockLeadStorer
    utmStore    mocks.MockUTMStorer
    asesorStore mocks.MockAsesorStorer
    rr          *roundrobin.RoundRobin[models.Asesor]

    commService *handlers.CommunicationService
)

func TestMain(t *testing.M) {
    leadStore = mocks.MockLeadStorer{}
    utmStore = mocks.MockUTMStorer{}
    sourceStore := mocks.MockSourceStorer{}
    asesorStore = *mocks.NewMockAsesorStore()

    leadStore.Mock()
    utmStore.Mock()
    asesorStore.Mock()

    asesores, _ := asesorStore.GetAll()
    rr = roundrobin.New(asesores)

    asesorService := handlers.NewAsesorService(&asesorStore, &leadStore, rr)

    commService = &handlers.CommunicationService{
        RoundRobin: rr,
        Logger: slog.Default(),
        Leads: &leadStore,
        Utms: &utmStore,
        Source: &sourceStore,
    }

    leadHandler = handlers.NewLeadHandler(&leadStore) 
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
            println(req.Method, req.URL.String(), string(response))
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
