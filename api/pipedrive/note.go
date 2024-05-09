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

@%s
`

func (p *Pipedrive) addNote(c *models.Communication, dealId uint32) (*Note, error){
    content := fmt.Sprintf(noteContent, 
        c.Fuente,
        c.Fecha,
        c.Link,
        c.Propiedad.Titulo,
        c.Propiedad.Link,
        c.Propiedad.Precio,
        c.Propiedad.Ubicacion,
        c.Propiedad.Tipo,
        c.Asesor.Name,
    )

    payload := &CreateNote{
        Content: content,
        DealId: dealId,
    }
    fmt.Printf("payload: %v", payload)

    var note Note
    err := p.makeRequest("POST", "notes", payload, &note)

    if err != nil{
        return nil, err
    }
    return &note, nil
}
