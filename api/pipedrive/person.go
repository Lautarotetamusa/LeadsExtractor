package pipedrive

import (
	"fmt"
	"leadsextractor/models"
)

type PersonChannel struct{
    Value   string  `json:"value"`
    Primary bool    `json:"primary"`
}

type Person struct{
    Id      uint32 `json:"id"`
    Name    string `json:"name"`
    Phone   []PersonChannel `json:"phone"`
    Email   []PersonChannel `json:"email"`
    // Owner   *User           `json:"owner_id"`
}

type FieldOption struct{
    Id      uint32 `json:"id"`
    Label   string `json:"label"`
}

//TODO: Que esto no esté hardcodeado

//Los atributos custom de pipedrive tienen un id para identificarlos
var customFields = map[string]string{
    "fuente": "0a551c8c09663ce803d924f036af12c3cc6b8b73",
    "fecha lead": "a7b3035eaea7ae5cb3aeab97f8f91748aa8e427b",
    "link": "180b604d295a05730ea6f453a384a3dc78bb108c",
    "zona": "41a4e5fc09fbc6243074f02d9dc284b8f9b1a505",
    "ubicacion": "8274fd2ca686df722bb1e78c5479acaba1067058",
    "precio": "80b2cb3150d88f04e61e442b76d942e636596274",
    "cotizacion": "982e5632726bfefe555bb5a7daea9a2bcbea1bf1",
};

var fuenteOptions = map[string]uint32{
    "vivanuncios": 21,
    "rebora": 22,
    "inmuebles24": 23,
    "lamudi": 24,
    "casasyterrenos": 25,
    "propiedades": 26,
    "whatsapp": 74,
    "ivr": 75,
};

func (p *Pipedrive) createPerson(c *models.Communication, ownerId uint32) (*Person, error){
    payload := map[string]interface{}{
        "name": c.Nombre,
        "owner_id": ownerId,
        "visible_to": "3", //visible para todos
        "phone": PersonChannel{
            Value: c.Telefono,
            Primary: true,
        },
        customFields["fuente"]: fuenteOptions[c.Fuente],
        customFields["fecha lead"]: c.FechaLead,
        customFields["link"]: c.Propiedad.Link.String,
        //La zona no la pongo porque es un campo que tiene valores opcionales (tendría que cargar esta zona como opcion)
        customFields["ubicacion"]: c.Propiedad.Ubicacion.String,
        customFields["precio"]: c.Propiedad.Precio.String,
        customFields["cotizacion"]: c.Cotizacion,
    }

    if c.Email.Valid {
        payload["email"] = PersonChannel{
            Value: c.Email.String,
            Primary: true,
        }
    }

    var person Person
    err := p.makeRequest("POST", "persons", payload, &person)

    if err != nil{
        return nil, err
    }
    return &person, nil
}

func (p *Pipedrive) MergePersons(aId uint32, bId uint32) (*Person, error){
    url := fmt.Sprintf("persons/%d/merge", aId)

    type Payload struct{
        Id  uint32 `json:"merge_with_id"`
    }
    payload := Payload{bId}
    var person Person

    if err := p.makeRequest("PUT", url, payload, &person); err != nil{
        return nil, err
    }

    return &person, nil
}

func (p *Pipedrive) GetPersonByNumber(number string) (*Person, error){
    url := fmt.Sprintf("persons/search?term=%s&fields=%s", number, "phone")

    var data struct{
        Items []struct{
            Item Person `json:"item"`
        } `json:"items"`
    }
    err := p.makeRequest("GET", url, nil, &data)

    if err != nil{
        return nil, err
    }
    if len(data.Items) == 0{
        return nil, fmt.Errorf("la persona con telefono %s no existe", number)
    }

    return &data.Items[0].Item, nil
}

func (p *Pipedrive) getField(id string) (*FieldOption, error){
    var field *FieldOption

    url := fmt.Sprintf("personFields/%s", id)
    err := p.makeRequest("GET", url, nil, field)

    if err != nil{
        return nil, fmt.Errorf("el field: %s, no existe", id)
    }

    return field, nil
}
