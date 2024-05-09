package pipedrive

import (
	"fmt"
	"leadsextractor/models"
)

type Deal struct{
    Id          uint32  `json:"id"`
    Title       string  `json:"title"`
    Value       int32   `json:"value"`
    Currency    string  `json:"currency"`
    User        User    `json:"user_id"`
    //VisibleTo   string  `json:"visible_to"` 
}

func (p *Pipedrive) searchPersonDeal(personId uint32, userId uint32) (*Deal, error){
    url := fmt.Sprintf("persons/%d/deals", personId)

    var deals []Deal
    err := p.makeRequest("GET", url, nil, &deals)

    if err != nil{
        return nil, err
    }
    if len(deals) == 0{
        return nil, fmt.Errorf("La asesor con id: %d no tiene ningun trato con la persona: %d", userId, personId)
    }

    for _, deal := range deals{
        if deal.User.Id == userId{
            return &deal, nil
        }
    }

    return nil, fmt.Errorf("La asesor con id: %d no tiene ningun trato con la persona: %d", userId, personId)
}

func (p *Pipedrive) createDeal(c *models.Communication, userId uint32, personId uint32) (*Deal, error){
    title := c.Propiedad.Titulo
    if title == "" {
        title = "Trato con " + c.Nombre
    }

    payload := map[string]interface{}{
        "title": title,
        "user_id": userId,
        "person_id": personId,
        "value": c.Propiedad.Precio,
        "currency": "MXN",
        "visible_to": "3", //Para todos
    }

    var deal Deal
    err := p.makeRequest("POST", "deals", payload, &deal)

    if err != nil{
        return nil, err
    }
    return &deal, nil
}
