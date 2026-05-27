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
}

type CreatePerson struct {
	Name      string         `json:"name"`
	OwnerId   uint32         `json:"owner_id"`
	VisibleTo string         `json:"visible_to"`
	Phone     PersonChannel  `json:"phone"`
	Email     *PersonChannel `json:"email,omitempty"`

	// Custom fields
	Fuente       uint32            `json:"0a551c8c09663ce803d924f036af12c3cc6b8b73,omitempty"`
	FechaLead    string            `json:"a7b3035eaea7ae5cb3aeab97f8f91748aa8e427b,omitempty"`
	Link         string            `json:"180b604d295a05730ea6f453a384a3dc78bb108c,omitempty"`
	Ubicacion    string            `json:"8274fd2ca686df722bb1e78c5479acaba1067058,omitempty"`
	Precio       string            `json:"80b2cb3150d88f04e61e442b76d942e636596274,omitempty"`
	AreaTotal    string            `json:"a9ca1dc337fc2fc24d2c5cb453426d03ad213809,omitempty"`
	AreaCubierta string            `json:"71f8984cf0276d7595bc06c16ace472a8e8c1175,omitempty"`
	UtmChannel   models.NullString `json:"d1004a59705fc22a31ce5a106bd749255945aae7"`
	UtmAd        string            `json:"7585f2a2444e77410a9c9339879294269ab51607,omitempty"`
	UtmCampaign  string            `json:"d594db0232426e8890c66c1b27d182358d8c5da4,omitempty"`
	UtmMedium    string            `json:"f1ae15f0f9f0800e30dbd0c5dfa192b3ea8da3f4,omitempty"`
	UtmSource    string            `json:"e50923b345eb0ffc139e9752ddc92fa843559a78,omitempty"`
	Mensaje      models.NullString `json:"a7032dfa893a197e2cf149791e6285e577c5641d"`
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

func (p *Pipedrive) createPerson(c *models.Communication, ownerId uint32) (*Person, error) {
	payload := CreatePerson{
		Name:      c.Nombre,
		OwnerId:   ownerId,
		VisibleTo: "3",
		Phone: PersonChannel{
			Value:   c.Telefono.String(),
			Primary: true,
		},
		Fuente:       fuenteOptions[c.Fuente],
		FechaLead:    c.FechaLead,
		Link:         c.Propiedad.Link.String,
		Ubicacion:    c.Propiedad.Ubicacion.String,
		Precio:       c.Propiedad.Precio.String,
		Mensaje:      c.Message,
		AreaTotal:    c.Propiedad.TotalArea,
		AreaCubierta: c.Propiedad.CoveredArea,
		UtmChannel:   c.Utm.Channel,
		UtmAd:        c.Utm.Ad.String,
		UtmCampaign:  c.Utm.Campaign.String,
		UtmMedium:    c.Utm.Medium.String,
		UtmSource:    c.Utm.Source.String,
	}

	if c.Email.Valid {
		payload.Email = &PersonChannel{
			Value:   c.Email.String,
			Primary: true,
		}
	}

	var person Person
	if err := p.makeRequest("POST", "persons", payload, &person); err != nil {
		return nil, err
	}
	return &person, nil
}

func (p *Pipedrive) MergePersons(aId uint32, bId uint32) (*Person, error) {
	url := fmt.Sprintf("persons/%d/merge", aId)

	type Payload struct {
		Id uint32 `json:"merge_with_id"`
	}
	var person Person

	if err := p.makeRequest("PUT", url, Payload{bId}, &person); err != nil {
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
	if err := p.makeRequest("GET", url, nil, &data); err != nil {
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
	return p.makeRequest("PUT", fmt.Sprintf("persons/%d", personId), payload, nil)
}
