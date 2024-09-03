package pkg

import (
	"encoding/json"
	"fmt"
	"leadsextractor/models"
	"leadsextractor/numbers"
	"net/http"
	"strings"
)

func (s *Server) reciveIVR(w http.ResponseWriter, r *http.Request) error {
    s.logger.Debug("IVR recived")
    msidsn := r.URL.Query().Get("msidsn")
    if msidsn == "" {
        s.logger.Error("Recive IVR without msidsn field")
        return fmt.Errorf("msidsn is required")
    }
    var phone *numbers.PhoneNumber 

    // Parsear telefono
    phone, err := numbers.NewPhoneNumber(msidsn)
    if err != nil {
        // Si el numero no es mexicano va a llegar con un 52 adelante igual por ejemplo 525493415854220
        // Removemos el '52'
        msidsn = msidsn[2:]
        phone, err = numbers.NewPhoneNumber(msidsn)
        if err != nil {
            return err
        }
    }
    c := models.Communication{
        Fuente: "ivr",
        Nombre: phone.String(),
        Telefono: *phone,
    }

    if err := s.NewCommunication(&c); err != nil {
        return err
    }

    // Parseamos el numero para que el IVR lo pueda conectar
    // TODO: Cambiar el tipo de dato de PhoneNumber 
    strPhone := strings.Replace(c.Asesor.Phone.String(), "+", "", 1)
    strPhone = strings.Replace(strPhone, "521", "52", 1)
    strPhone = strings.Replace(strPhone, "549", "54", 1)
    c.Asesor.Phone = numbers.PhoneNumber(strPhone)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(c.Asesor)
    return nil
}
