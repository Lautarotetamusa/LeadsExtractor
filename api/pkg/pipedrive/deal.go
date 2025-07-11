package pipedrive

import (
	"fmt"
	"leadsextractor/models"
)

type Deal struct {
	Id       uint32 `json:"id"`
	Title    string `json:"title"`
	Value    uint64 `json:"value"`
	Currency string `json:"currency"`
	User     User   `json:"user_id"`
	//VisibleTo   string  `json:"visible_to"`
	Person *Person    `json:"person_id"`
	Status DealStatus `json:"status"`

    // If the deal its closed, the stage_id stays in the same as before
    StageId     int `json:"stage_id"`
    PipelineId  int `json:"pipeline_id"`
}

type CreateDeal struct {
	Title     string `json:"title"`
	Value     string `json:"value"`
	Currency  string `json:"currency"`
	UserId    uint32 `json:"user_id"`
	PersonId  uint32 `json:"person_id"`
	VisibleTo string `json:"visible_to"`
}

type UpdateDeal struct {
	Status DealStatus `json:"status"`
    StageId int `json:"stage_id"`
}

// Buscamos para la persona con id personId una trato con el asesor con id userId
func (p *Pipedrive) SearchPersonDeal(personId uint32, userId uint32) (*Deal, error) {
	url := fmt.Sprintf("persons/%d/deals", personId)

	var deals []Deal

	err := p.makeRequest("GET", url, nil, &deals)

	if err != nil {
		return nil, err
	}

	if len(deals) == 0 {
		return nil, fmt.Errorf("el asesor con id: %d no tiene ningun trato con la persona: %d", userId, personId)
	}

	for _, deal := range deals {
		if deal.User.Id == userId {
			return &deal, nil
		}
	}

	return nil, fmt.Errorf("el asesor con id: %d no tiene ningun trato con la persona: %d", userId, personId)
}

func (p *Pipedrive) createDeal(c *models.Communication, userId uint32, personId uint32) (*Deal, error) {
	title := c.Propiedad.Titulo.String
	if title == "" {
		title = "Trato con " + c.Nombre
	}

	payload := CreateDeal{
		Title:     title,
		UserId:    userId,
		PersonId:  personId,
		Value:     c.Propiedad.Precio.String,
		Currency:  "MXN",
		VisibleTo: "3", //Para todos
	}

	var deal Deal
	err := p.makeRequest("POST", "deals", payload, &deal)

	if err != nil {
		return nil, err
	}
	return &deal, nil
}

func (p *Pipedrive) updateDeal(dealId uint32, u *UpdateDeal) (*Deal, error) {
	var deal Deal
	err := p.makeRequest("PUT", fmt.Sprintf("deals/%d/", dealId), u, &deal)

	if err != nil {
		return nil, err
	}
	return &deal, nil
}

func (p *Pipedrive) ReasignDeal(dealId uint32, newUserId uint32) (*Deal, error) {
	payload := map[string]interface{}{
		"user_id": newUserId,
	}

	var deal Deal
	err := p.makeRequest("PUT", fmt.Sprintf("deals/%d/", dealId), payload, &deal)

	if err != nil {
		return nil, err
	}
	return &deal, nil
}
