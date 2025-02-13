package handlers_test

import (
	"leadsextractor/models"
	"leadsextractor/store"
)


type mockCommStorer struct {
    comms []models.Communication
}

const mockComm = `{
    "fuente": "inmuebles24",
    "fecha_lead": "2025-02-11",
    "lead_id": "471275112",
    "nombre": "Daniel",
    "link": "https://www.inmuebles24.com/panel/interesados/221934608",
    "telefono": "526865118873",
    "email": "gomez.albanesz@gmail.com",
    "message": "Hola! , muchas gracias por tu mensaje, tu asesor asignado es: Hernan \u200d\u2642\ufe0f, En unos momentos se pone en contacto contigo. Muchas gracias por la espera! Mientras tanto, quiero invitarte a que escuches nuestro podcast en Spotify \u201dvas a invertir en bienes ra\u00edces? - lo que nadie te dice, \u00f3 a cotizar tu casa en l\u00ednea en 2 clics! \n\nCotiz\u00e1 tu casa en l\u00ednea!: https://rebora.com.mx/cotiza-en-linea/\n\nEscucha nuestro podcast: https://open.spotify.com/show/5ylqI4dl1rmgwYvMXVKvnw?si=OhyfK4ztR5-5KaS9dPy9Uw",
    "estado": "",
    "cotizacion": "",
    "propiedad": {
        "id": "143664764",
        "titulo": "Casa en venta en Bugambilias  Dise\u00f1o moderno que genera plusval\u00eda ",
        "link": "https://www.inmuebles24.com/propiedades/-143664764.html",
        "precio": "18550000",
        "ubicacion": "P.\u00ba de Las Hortensias, Bugambilias, 45100 Zapopan, Jal.",
        "tipo": "Casa",
        "bedrooms": "",
        "bathrooms": "",
        "total_area": "",
        "covered_area": "",
        "municipio": "Zapopan"
    },
    "busquedas": {
        "zonas": "",
        "tipo": "",
        "presupuesto": ", ",
        "cantidad_anuncios": null,
        "contactos": null,
        "inicio_busqueda": null,
        "total_area": "",
        "covered_area": "",
        "banios": "",
        "recamaras": ""
    },
    "contact_id": "221934608",
    "utm": {
        "utm_channel": "whatsapp"
    }
}`

func (s *mockCommStorer) Insert(c *models.Communication, source *models.Source) error {
    return nil
}

func (s *mockCommStorer) GetAll(params *store.QueryParam) ([]models.Communication, error) {
    return s.comms, nil
}

func (s *mockCommStorer) GetDistinct(params *store.QueryParam) ([]models.Communication, error) {
    return nil, store.NewErr("does not exists", store.StoreNotFoundErr)
}

func (s *mockCommStorer) Count(params *store.QueryParam) (int, error) {
    return len(s.comms), nil
}

func (s *mockCommStorer) Exists(params *store.QueryParam) bool {
    return true
}
