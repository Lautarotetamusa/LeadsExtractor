package pipedrive

import (
	"fmt"
	"leadsextractor/models"
)

type CreateNote struct{
    Content     string  `json:"content"`
    DealId      uint32  `json:"deal_id"`
}

type Note struct{
}

const noteContent = `
fuente: %s
fecha: %s
link lead: %s
titulo: %s
link propiedad: %s
precio: %s
ubicacion: %s
tipo: %s
bedrooms: %s,
bathrooms: %s,
total_area: %s,
covered_area: %s

@%s
`

func (p *Pipedrive) addNote(c *models.Communication, dealId uint32) (*Note, error){
    content := fmt.Sprintf(noteContent, 
        c.Fuente,
        c.Fecha,
        c.Link,
        c.Propiedad.Titulo.String,
        c.Propiedad.Link.String,
        c.Propiedad.Precio.String,
        c.Propiedad.Ubicacion.String,
        c.Propiedad.Tipo.String,
        c.Propiedad.Bedrooms,
        c.Propiedad.Bathrooms,
        c.Propiedad.TotalArea,
        c.Propiedad.CoveredArea,
        c.Asesor.Name,
    )

    payload := &CreateNote{
        Content: content,
        DealId: dealId,
    }

    var note Note
    err := p.makeRequest("POST", "notes", payload, &note)

    if err != nil{
        return nil, err
    }
    return &note, nil
}
