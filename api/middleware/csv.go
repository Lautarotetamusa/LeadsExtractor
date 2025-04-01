package middleware

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strings"

	"github.com/gocarina/gocsv"
)

type CSVHandler[T any] struct {
	limit       int
    fileName    string
    contextKey  string
    requiredHeaders []string
    items       []T
}

const defaultLimit = 100

func NewCSVHandler[T any](fileName, contextKey string, items []T) *CSVHandler[T] {
    return &CSVHandler[T]{
        fileName: fileName,
        contextKey: contextKey,
        limit: defaultLimit,
        items:  items,
        requiredHeaders: getStructHeaders(items),
    }
}

func (c *CSVHandler[T]) WithItems(items []T) *CSVHandler[T] {
    c.items = items
    return c
}

func (c *CSVHandler[T]) WithLimit(limit int) *CSVHandler[T] {
    c.limit = limit
    return c
}

func (c *CSVHandler[T]) CSVMiddleware(next http.Handler) http.Handler {
	return handleErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		// Parse multipart form (adjust max memory as needed)
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return err
		}

		file, _, err := r.FormFile(c.fileName)
		if err != nil {
			return fmt.Errorf("error retrieving %s: %v", c.fileName, err)
		}
		defer file.Close()

		csvReader := csv.NewReader(file)
		headers, err := csvReader.Read()
		if err != nil {
			return fmt.Errorf("error reading CSV headers: %v", err)
		}

		if err := c.validateHeaders(headers); err != nil {
			return err
		}

		// Reset file reader, supose this will not going to fail
		file.Seek(0, io.SeekStart)

		// Unmarshal CSV data
		// var comms []models.Communication
		if err := gocsv.UnmarshalMultipartFile(&file, &c.items); err != nil {
			return fmt.Errorf("error parsing CSV data: %v", err)
		}

        // for get the len of the items
        // items, _ := c.items.([]interface{})

		// Validate maximum entries
		if len(c.items) > c.limit {
			return fmt.Errorf("the csv limit its %d, rows count: %d", c.limit, len(c.items))
		}

		ctx := context.WithValue(r.Context(), c.contextKey, c.items)
		next.ServeHTTP(w, r.WithContext(ctx))
        return nil
	})
}

func getStructHeaders(d interface{}) []string {
    var headers []string
    val := reflect.ValueOf(d)

    if val.Kind() == reflect.Ptr {
        val = val.Elem()
    }

    t := val.Type()

    if t.Kind() == reflect.Slice {
        t = t.Elem()
        if t.Kind() == reflect.Ptr {
            t = t.Elem()
        }
    }

    if t.Kind() != reflect.Struct {
        return headers
    }

    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        csvTag := field.Tag.Get("csv")

        // Obtener el tipo subyacente (manejar punteros)
        fieldType := field.Type
        if fieldType.Kind() == reflect.Ptr {
            fieldType = fieldType.Elem()
        }

        // Procesar campos anidados
        if fieldType.Kind() == reflect.Struct {
            nestedPrefix := csvTag // Ej: "ubication"
            if nestedPrefix != "" {
                nestedPrefix += "." // Ej: "ubication."
            }

            nestedVal := reflect.New(fieldType).Elem().Interface()
            nestedHeaders := getStructHeaders(nestedVal)

            // Agregar prefijo a los headers anidados
            for _, h := range nestedHeaders {
                headers = append(headers, nestedPrefix+h) // Ej: "ubication.address"
            }
        } else if csvTag != "" {
            headers = append(headers, csvTag) // Campos no-estructos
        }
    }

    return headers
}

// Example usage in router setup
// func main() {
// 	r := mux.NewRouter()
// 	r.Handle("/upload", CommunicationsCSVMiddleware(
// 		http.HandlerFunc(uploadHandler),
// 	))
// 	http.ListenAndServe(":8080", r)
// }
// func uploadHandler(w http.ResponseWriter, r *http.Request) {
// 	comms, ok := GetCommunicationsFromContext(r)
// 	if !ok {
// 		http.Error(w, "Communications data missing", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Process communications...
// 	fmt.Fprintf(w, "Successfully processed %d communications", len(comms))
// }

func (c *CSVHandler[T]) validateHeaders(headers []string) error {
	missingFields := make([]string, 0)
	for _, header := range c.requiredHeaders {
		if !slices.Contains(headers, header) {
			missingFields = append(missingFields, fmt.Sprintf("'%s'", header))
		}
	}
	if len(missingFields) > 0 {
		return fmt.Errorf("fields %s are required", strings.Join(missingFields, ", "))
	}
	return nil
}

type HandlerErrFunc func(http.ResponseWriter, *http.Request) error
func handleErrorFunc(fn HandlerErrFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			json.NewEncoder(w).Encode(map[string]any{
				"success": false,
				"error":   err.Error(),
			})
		}
	})
}
