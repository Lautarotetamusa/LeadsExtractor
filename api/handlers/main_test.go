package handlers_test

import (
	"leadsextractor/handlers"
	"net/http"
	"testing"
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

var utmHandler  handlers.UTMHandler
var leadHandler handlers.LeadHandler

func TestMain(t *testing.M) {
    leadHandler = *handlers.NewLeadHandler(mockLeadStorer{}) 
    utmHandler = *handlers.NewUTMHandler(mockUTMStorer{})

    t.Run()
}

