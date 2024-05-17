package whatsapp

import (
	"fmt"
	"testing"
)

func TestSendMessage(t *testing.T){
    w := NewWhatsapp(
        "EAAbDrjrkX5IBO19E0ZAlpHa8iTcm17YoZC0aCmC9yu2xLpOqvOmcsp3KbJmU8q94meZBhYSLs283AgaZANhiZBt37YqvOuxEg4KxAYIm7ShzHn3bPDKBkq7eBB3IHYSUPVc3LaMzswwd4pXfcRUVVZARLWQ074WZBfrvkAkkBs5Jm7ZBOZAtfCbbUP4OVrkUP1UNA",
        "193151663891728",
    )

    res, err := w.SendMessage("+5493415854220", "Hola! mensaje de prueba")
    if err != nil{
        t.Errorf("Error enviando el mensaje %s", err)
    }
    fmt.Printf("%v", res)
}
