package whatsapp

import (
	"fmt"
	"leadsextractor/models"
	"testing"
)

const phone = "+5493415854220"

func TestSendMessage(t *testing.T) {
	w := NewWhatsapp(
		"EAAbDrjrkX5IBO19E0ZAlpHa8iTcm17YoZC0aCmC9yu2xLpOqvOmcsp3KbJmU8q94meZBhYSLs283AgaZANhiZBt37YqvOuxEg4KxAYIm7ShzHn3bPDKBkq7eBB3IHYSUPVc3LaMzswwd4pXfcRUVVZARLWQ074WZBfrvkAkkBs5Jm7ZBOZAtfCbbUP4OVrkUP1UNA",
		"193151663891728",
	)

	res, err := w.SendMessage(phone, "Hola! mensaje de prueba")
	if err != nil {
		t.Errorf("Error enviando el mensaje %s", err)
	}
	fmt.Printf("%v", res)
}

func TestSendTemplate(t *testing.T) {
	w := NewWhatsapp(
		"EAAbDrjrkX5IBO19E0ZAlpHa8iTcm17YoZC0aCmC9yu2xLpOqvOmcsp3KbJmU8q94meZBhYSLs283AgaZANhiZBt37YqvOuxEg4KxAYIm7ShzHn3bPDKBkq7eBB3IHYSUPVc3LaMzswwd4pXfcRUVVZARLWQ074WZBfrvkAkkBs5Jm7ZBOZAtfCbbUP4OVrkUP1UNA",
		"193151663891728",
	)

	a := &models.Asesor{
		Name:  "test",
		Phone: "+5493415854220",
		Email: "lautarotetamusa@gmail.com",
	}

	res, err := w.SendResponse(phone, a)
	if err != nil {
		t.Errorf("Error enviando el mensaje %s", err)
	}
	fmt.Printf("%v", res)
}

func TestSendDocument(t *testing.T) {
	w := NewWhatsapp(
		"EAAbDrjrkX5IBO19E0ZAlpHa8iTcm17YoZC0aCmC9yu2xLpOqvOmcsp3KbJmU8q94meZBhYSLs283AgaZANhiZBt37YqvOuxEg4KxAYIm7ShzHn3bPDKBkq7eBB3IHYSUPVc3LaMzswwd4pXfcRUVVZARLWQ074WZBfrvkAkkBs5Jm7ZBOZAtfCbbUP4OVrkUP1UNA",
		"193151663891728",
	)

	url := "https://www.frro.utn.edu.ar/repositorio/catedras/quimica/5_anio/orientadora1/monograias/matich-redesneuronales.pdf"
	res, err := w.Send(NewDocumentPayload(phone, url, "archivo de prueba", "test"))
	if err != nil {
		t.Errorf("Error enviando el mensaje %s", err)
	}
	fmt.Printf("%v", res)
}

func TestSendMsgAsesor(t *testing.T) {
	w := NewWhatsapp(
		"EAAbDrjrkX5IBO19E0ZAlpHa8iTcm17YoZC0aCmC9yu2xLpOqvOmcsp3KbJmU8q94meZBhYSLs283AgaZANhiZBt37YqvOuxEg4KxAYIm7ShzHn3bPDKBkq7eBB3IHYSUPVc3LaMzswwd4pXfcRUVVZARLWQ074WZBfrvkAkkBs5Jm7ZBOZAtfCbbUP4OVrkUP1UNA",
		"193151663891728",
	)

	c := &models.Communication{
		Fuente:    "inmuebles24",
		FechaLead: "2024-04-07",
		ID:        "461161340",
		Fecha:     "2024-04-08",
		Nombre:    "Lautaro",
		Link:      "https://www.inmuebles24.com/panel/interesados/198059132",
		Telefono:  "5493415854220",
		Email:     "cornejoy369@gmail.com",
		Propiedad: models.Propiedad{
			ID:        "a78c1555-f684-4de7-bbf1-a7288461fe51",
			Titulo:    "Casa en venta en El Cielo Country Club Incre\u00edble dise\u00f1o y amplitud",
			Link:      "",
			Precio:    "16117690",
			Ubicacion: "cielo country club",
			Tipo:      "Casa",
		},
	}

	res, err := w.SendMsgAsesor(phone, c, true)
	if err != nil {
		t.Errorf("Error enviando el mensaje %s", err)
	}
	fmt.Printf("%v", res)

	res, err = w.SendMsgAsesor(phone, c, false)
	if err != nil {
		t.Errorf("Error enviando el mensaje %s", err)
	}
	fmt.Printf("%v", res)
}
