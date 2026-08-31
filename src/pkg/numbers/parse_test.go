package numbers

import (
	"testing"
)

// Los que estan como "" no deben pasar
var expecteds = map[string]string{
	"258192":          "",
	"12345":           "",
	"3314156138":      "+5213314156138",
	"5213314156138":   "+5213314156138",
	"+527441036217":   "+5217441036217",
	"+549 3415854220": "+5493415854220",
	"+543415854220":   "+5493415854220",
	"+5491234":        "",
	"3415854220":      "+5213415854220",
	"+523324944591":   "+5213324944591",
	"523327919473":    "+5213327919473",
	"523411234567":    "+5213411234567",
	"3344556677":      "+5213344556677",
	"+524151510540":   "+5214151510540",
	"3415725421":      "+5213415725421",
	"5493415854220":   "+5493415854220",
}

func TestWppNumber(t *testing.T) {
	for number, expected := range expecteds {
		formatted, err := NewPhoneNumber(number)
		if expected != "" {
			if err != nil {
				t.Errorf("cannot parse %s %s", number, err.Error())
				continue
			}
			if formatted.String() != expected {
				t.Errorf(`expected=%s, recived=%s`, expected, *formatted)
			}
		} else {
			if err == nil {
				t.Errorf("number %s must not be parsed. parsed as %s", number, formatted)
				continue
			}
		}
	}
}
