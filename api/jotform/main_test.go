package jotform

import (
	"fmt"
	// "leadsextractor/models"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestObtainPdf(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

    jotform := NewJotform(
        os.Getenv("JOTFORM_API_KEY"),
        "http://localhost:8081",
    )
    
    form := jotform.AddForm(os.Getenv("JOTFORM_FORM_ID"))
    submissionId := "6017823550128303377"
    fmt.Printf("%#v\n", form)

    res, err := jotform.ObtainPdf(submissionId, form)
    if err != nil {
        t.Error(err.Error())
        fmt.Println(err.Error())
    }
    fmt.Printf("%#v\n", res)
}

/*
func TestSubmitForm(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

    jotform := NewJotform(
        os.Getenv("JOTFORM_API_KEY"),
        "http://localhost:8081",
    )
    form := jotform.AddForm(os.Getenv("JOTFORM_FORM_ID"))

	c := &models.Communication{
		Fuente:    "inmuebles24",
		FechaLead: "2024-04-07",
		Fecha:     "2024-04-08",
		Nombre:    "Lautaro",
		Link:      "https://www.inmuebles24.com/panel/interesados/198059132",
		Telefono:  "5493415854220",
        Email:     models.NullString{String: "cornejoy369@gmail.com"},
        Asesor:     models.Asesor{
            Email: "test@test.com",
            Name:   "test",
            Phone: "5493415854220",
        },
		Propiedad: models.Propiedad{
            PortalId:   models.NullString{String: "a78c1555-f684-4de7-bbf1-a7288461fe51"},
            Titulo:     models.NullString{String: "Casa en venta en El Cielo Country Club Incre\u00edble dise\u00f1o y amplitud"},
            Link:       models.NullString{String: ""},
            Precio:     models.NullString{String: "16117690"},
            Ubicacion:  models.NullString{String: "cielo country club"},
            Tipo:       models.NullString{String: "Casa"},
		},
	}

    url, err := jotform.GetPdf(c, form)
    if err != nil {
        t.Error(err.Error())
        fmt.Println(err.Error())
    }
    fmt.Println(url)
}
*/
