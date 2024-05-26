package whatsapp

import (
	"database/sql"
	"fmt"
	"leadsextractor/models"
	"log/slog"
	"os"
	"testing"

	"github.com/lmittmann/tint"
)

const phone = "+5493415854220"

func TestSendMsgAsesor(t *testing.T) {
	w := NewWhatsapp(
		"EAAbDrjrkX5IBO1rbBqzz5SnpHLENnMTOY45DN7Y39gSRyrLxcfQmyBNE8ShQHMUTOtXnaZBZAWtrA6Scx6H6cdQCZAMSPsj3KVJBcExm3jFyeROA3FRwPGn08GFNkZA8ZCPIMy8BOPZCOUSv4Ou66PtVscYts0kAe5UsjVe9ufSw2Kywv8XdrZBpbumdUmflcvB",
		"193151663891728",
        slog.New(tint.NewHandler(os.Stderr, nil)),
	)

	c := &models.Communication{
		Fuente:    "inmuebles24",
		FechaLead: "2024-04-07",
		ID:        "461161340",
		Fecha:     "2024-04-08",
		Nombre:    "Lautaro",
		Link:      "https://www.inmuebles24.com/panel/interesados/198059132",
		Telefono:  "5493415854220",
        Email:     sql.NullString{String: "cornejoy369@gmail.com"},
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
