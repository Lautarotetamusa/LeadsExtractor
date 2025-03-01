package handlers_test

import (
	"fmt"
	"leadsextractor/handlers"
	"net/http"
	"strings"
	"testing"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/gorilla/mux"
)

func TestPropCRUD(t *testing.T) {
	propsListRes := handlers.NewDataResponse(propStore.Props())

    valid := map[string]any {
        "title": "Cozy House in the Suburbs",
        "price": "300000",
        "currency": "EUR",
        "description":  "A charming 2-bedroom house with a large garden.",
        "type": "house",
        "operation_type": "sale",
        "antiquity": 10,
        "rooms": 2,
        "m2_total": 200,
        "m2_covered": 150,
        "state": "Jalisco",
        "municipality": "Zapopan",
        "colony": "Zapopan centro",
        "street": "Urquiza",
        "number": "1159",
        "zip_code": "2000",
        "images": []map[string]any{
            {"url": "https://domain.com/image1",},
            {"url": "https://domain.com/image2",},
        },
    };
    expected := map[string]any {
        "success": true,
        "data": map[string]any{
            "id": 4,
            "title": "Cozy House in the Suburbs",
            "price": "300000",
            "currency": "EUR",
            "description":  "A charming 2-bedroom house with a large garden.",
            "operation_type": "sale",
            "bathrooms": nil,
            "half_bathrooms": nil,
            "parking_lots": nil,
            "type": "house",
            "antiquity": 10,
            "rooms": 2,
            "m2_total": 200,
            "m2_covered": 150,
            "video_url": nil,
            "virtual_route": nil,
            "updated_at": time.Time{},
            "created_at": time.Time{},
            "state": "Jalisco",
            "municipality": "Zapopan",
            "colony": "Zapopan centro",
            "neighborhood": nil,
            "street": "Urquiza",
            "number": "1159",
            "zip_code": "2000",
            "images": []map[string]any{
                {
                    "id": 1,
                    "url": "https://domain.com/image1",
                },
                {
                    "id": 2,
                    "url": "https://domain.com/image2",
                },
            },
        },
    };

    // without the image with id 1 and with the image with id 3. What i do in the tests
    expectedImages := []map[string]any{  
        { "id": 2, "url": "https://domain.com/image2", }, 
        { "id": 3, "url": "https://test.com/image3", }, 
    }

	tests := []APITestCase{
		{"get all", "GET", "/property", "", nil, http.StatusOK, Stringify(propsListRes)},
		// Create
		{"no pass body", "POST", "/property", "", nil, http.StatusBadRequest, ""},
        {"invalid price", "POST", "/property", `{"price": 99.4}`, nil, http.StatusBadRequest, `*number into Go struct field PortalProp.price of type string*`},
        {"invalid operation type", "POST", "/property", `{"operation_type": "invalid"}`, nil, http.StatusBadRequest, `*'OperationType' failed*`},
		{"create", "POST", "/property", Stringify(valid), nil, http.StatusCreated, `*property created successfully*`},
		{"get created", "GET", "/property/4", "", nil, http.StatusOK, Stringify(expected)},
		// Get
		{"get one", "GET", "/property/1", "", nil, http.StatusOK, Stringify(handlers.NewDataResponse(propStore.Props()[0]))},
		{"get one no exists", "GET", "/property/999", "", nil, http.StatusNotFound, "*does not exists*"},
        // Delete image
		{"delete image no exists", "DELETE", "/property/4/image/999", "", nil, http.StatusNotFound, "*does not exists*"},
		{"delete image", "DELETE", "/property/4/image/1", "", nil, http.StatusOK, "*deleted successfully*"},
        {"add invalid image", "POST", "/property/4/image", `[{"url": "thisisnotavalidurl"}]`, nil, http.StatusBadRequest, "*must have a valid url*"},
        {"add image", "POST", "/property/4/image", `[{"url": "https://test.com/image3"}]`, nil, http.StatusCreated, "*added successfully*"},
        {"get updated img", "GET", "/property/4/image", "", nil, http.StatusOK, Stringify(handlers.NewDataResponse(expectedImages))},
	}

    // Validate required fields
    requiredFields := []string{
        "state", "municipality", "colony", 
        "street", "number", "zip_code", 
        "title", "price", "description",
    }
    requiredTC := APITestCase{"", "POST", "/property", "", nil, http.StatusBadRequest, ""}
    for _, field := range requiredFields {
        requiredTC.Name = fmt.Sprintf("empty %s", field)
        requiredTC.Body = fmt.Sprintf(`{"%s": ""}`, field)

        // Convert the field name to CamelCase. zip_code => ZipCode
        field = strings.ReplaceAll(field, "_", " ")
        field = cases.Title(language.Spanish).String(field)
        field = strings.ReplaceAll(field, " ", "")

        requiredTC.WantResponse = fmt.Sprintf(`*'%s' failed*`, field)
        tests = append(tests, requiredTC)

        requiredTC.Name = fmt.Sprintf("no %s", field)
        requiredTC.Body = "{}"
        tests = append(tests, requiredTC)
    }

	router := mux.Router{}
	propHandler.RegisterRoutes(&router)
	for _, tc := range tests {
		Endpoint(t, &router, tc)
	}

	// upd := map[string]any{
	// 	"name": "juan carlos",
	// }
	// expected["name"] = "juan carlos"
	// updateRes := handlers.NewSuccessResponse(expected, "Lead actualizado correctamente")
	// updatedres := handlers.NewDataResponse(expected)
	// updateTC := APITestCase{"update", "PUT", "/property/5493415854220", Stringify(upd), nil, http.StatusCreated, Stringify(updateRes)}
	// Endpoint(t, &router, updateTC)
	// tc := APITestCase{"get updated", "GET", "/property/5493415854220", "", nil, http.StatusOK, Stringify(updatedres)}
	// Endpoint(t, &router, tc)
}
