package pipedrive

import (
	"fmt"
	"leadsextractor/models"
)

type PersonChannel struct {
	Value   string `json:"value"`
	Primary bool   `json:"primary"`
}

type Person struct {
	Id    uint32          `json:"id"`
	Name  string          `json:"name"`
	Phone []PersonChannel `json:"phone"`
	Email []PersonChannel `json:"email"`
	// Owner   *User           `json:"owner_id"`
}

type FieldOption struct {
	Id    uint32 `json:"id"`
	Label string `json:"label"`
}

//TODO: Que esto no esté hardcodeado

// Los atributos custom de pipedrive tienen un id para identificarlos
var customFields = map[string]string{
	"fuente":        "0a551c8c09663ce803d924f036af12c3cc6b8b73",
	"fecha lead":    "a7b3035eaea7ae5cb3aeab97f8f91748aa8e427b",
	"link":          "180b604d295a05730ea6f453a384a3dc78bb108c",
	"zona":          "41a4e5fc09fbc6243074f02d9dc284b8f9b1a505",
	"ubicacion":     "8274fd2ca686df722bb1e78c5479acaba1067058",
	"precio":        "80b2cb3150d88f04e61e442b76d942e636596274",
	"area_total":    "a9ca1dc337fc2fc24d2c5cb453426d03ad213809",
	"area_cubierta": "71f8984cf0276d7595bc06c16ace472a8e8c1175",
	"utm_channel":   "d1004a59705fc22a31ce5a106bd749255945aae7",
	"utm_ad":        "7585f2a2444e77410a9c9339879294269ab51607",
	"utm_campaign":  "d594db0232426e8890c66c1b27d182358d8c5da4",
	"utm_medium":    "f1ae15f0f9f0800e30dbd0c5dfa192b3ea8da3f4",
	"utm_source":    "e50923b345eb0ffc139e9752ddc92fa843559a78",
	"mensaje":       "a7032dfa893a197e2cf149791e6285e577c5641d",
}

var fuenteOptions = map[string]uint32{
	"vivanuncios":    21,
	"rebora":         22,
	"inmuebles24":    23,
	"lamudi":         24,
	"casasyterrenos": 25,
	"propiedades":    26,
	"whatsapp":       74,
	"ivr":            75,
}

func (p *Pipedrive) validatePersonCustomFields() error {
	for label, id := range customFields {
		field, err := p.getField(id)
		if err != nil {
			return err
		}
		if field.Label != label {
			return fmt.Errorf("el label del field con id %s no coincide %s != %s", id, field.Label, label)
		}
	}
	return nil
}

func (p *Pipedrive) createPerson(c *models.Communication, ownerId uint32) (*Person, error) {
	payload := map[string]interface{}{
		"name":       c.Nombre,
		"owner_id":   ownerId,
		"visible_to": "3", //visible para todos
		"phone": PersonChannel{
			Value:   c.Telefono.String(),
			Primary: true,
		},
		customFields["fuente"]:     fuenteOptions[c.Fuente],
		customFields["fecha lead"]: c.FechaLead,
		customFields["link"]:       c.Propiedad.Link.String,
		//La zona no la pongo porque es un campo que tiene valores opcionales (tendría que cargar esta zona como opcion)
		customFields["ubicacion"]: c.Propiedad.Ubicacion.String,
		customFields["precio"]:    c.Propiedad.Precio.String,

		customFields["mensaje"]: c.Message,

		customFields["area_total"]:    c.Propiedad.TotalArea,
		customFields["area_cubierta"]: c.Propiedad.CoveredArea,

		customFields["utm_channel"]:  c.Utm.Channel,
		customFields["utm_ad"]:       c.Utm.Ad.String,
		customFields["utm_campaign"]: c.Utm.Campaign.String,
		customFields["utm_medium"]:   c.Utm.Medium.String,
		customFields["utm_source"]:   c.Utm.Source.String,
	}

	if c.Email.Valid {
		payload["email"] = PersonChannel{
			Value:   c.Email.String,
			Primary: true,
		}
	}

	var person Person
	err := p.makeRequest("POST", "persons", payload, &person)

	if err != nil {
		return nil, err
	}
	return &person, nil
}

func (p *Pipedrive) MergePersons(aId uint32, bId uint32) (*Person, error) {
	url := fmt.Sprintf("persons/%d/merge", aId)

	type Payload struct {
		Id uint32 `json:"merge_with_id"`
	}
	payload := Payload{bId}
	var person Person

	if err := p.makeRequest("PUT", url, payload, &person); err != nil {
		return nil, err
	}

	return &person, nil
}

func (p *Pipedrive) GetPersonByNumber(number string) (*Person, error) {
	url := fmt.Sprintf("persons/search?term=%s&fields=%s", number, "phone")

	var data struct {
		Items []struct {
			Item Person `json:"item"`
		} `json:"items"`
	}
	err := p.makeRequest("GET", url, nil, &data)

	if err != nil {
		return nil, err
	}
	if len(data.Items) == 0 {
		return nil, fmt.Errorf("la persona con telefono %s no existe", number)
	}

	return &data.Items[0].Item, nil
}

// Solamente se puede actualizar el owner_id si el id del nuevo usuario tiene los permisos necesarios
// Los permisos se pueden ver en: https://api.pipedrive.com/v1/users/{user_id}/permissions
// Para poder cambiar el owner de una persona tiene que tener el permiso: can_modify_owner_for_people
func (p *Pipedrive) UpdatePersonOwner(personId uint32, newOwnerId uint32) error {
	payload := map[string]any{
		"owner_id": newOwnerId,
	}
	url := fmt.Sprintf("persons/%d", personId)

	err := p.makeRequest("PUT", url, payload, nil)

	if err != nil {
		return err
	}

	return nil
}

// El id no es el uuid de customFields si no un id int ej:9051
func (p *Pipedrive) getField(id string) (*FieldOption, error) {
	var field *FieldOption

	url := fmt.Sprintf("personFields/%s", id)
	err := p.makeRequest("GET", url, nil, field)

	if err != nil {
		return nil, fmt.Errorf("el field: %s, no existe", id)
	}

	return field, nil
}
