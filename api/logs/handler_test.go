package logs

import (
	"net/url"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestGreaterThanAndLowerThan(t *testing.T) {
    time_lt := "2024-12-05"
    time_gt := "2024-12-10"
    urlValues := url.Values{
        "time_gt": {time_gt},
        "time_lt": {time_lt},
    }

    expected := bson.M{
        "time": bson.M{
            "$gt": time_gt,
            "$lt": time_lt,
        },
    }

    filter := buildFilter(urlValues)
    if !reflect.DeepEqual(filter, expected) {
        t.Fatalf("expected: %#v. recived: %#v", expected, filter)
    }
}

// TODO: Testear que las fechas completas funcionen
// 2024-12-04T03:00:00.000Z
// 2024-12-09T03:00:00.000Z
// TODO: Testear completamente la funcion y no solo el buildFilter

// TODO: Parsear correctamente las fechas
// time_lt=2024-12-9  no devuelve nada

// TODO: Agregar mas tests
