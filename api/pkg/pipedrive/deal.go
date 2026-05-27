package pipedrive

import (
	"fmt"
	"leadsextractor/models"
	"strings"
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
	CampaignMkt string `json:"e3870475a190125348a7b55e113ffa7e3c4004da"`
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

var customDealFields = map[string]string{
	"campaign_mkt": "e3870475a190125348a7b55e113ffa7e3c4004da",
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

func findCampaignMkt(s string) (string, bool) {
	start := strings.Index(s, "[")
	if start == -1 {
		return "", false
	}
	end := strings.Index(s[start:], "]")
	if end == -1 {
		return "", false
	}

	// end es la posición relativa dentro de s[start:]; ajustamos al índice absoluto.
	return s[start+1 : start+end], true
}

// appendCampaignMkt agrega newMkt al valor existente separado por coma, evitando duplicados.
func appendCampaignMkt(existing, newMkt string) string {
	if existing == "" {
		return newMkt
	}
	for _, part := range strings.Split(existing, ",") {
		if strings.TrimSpace(part) == newMkt {
			return existing
		}
	}
	return existing + "," + newMkt
}

func (p *Pipedrive) updateDealCampaignMkt(deal *Deal, newMkt string) {
	updated := appendCampaignMkt(deal.CampaignMkt, newMkt)
	if updated == deal.CampaignMkt {
		return
	}
	payload := map[string]interface{}{
		customDealFields["campaign_mkt"]: updated,
	}
	if err := p.makeRequest("PUT", fmt.Sprintf("deals/%d/", deal.Id), payload, deal); err != nil {
		p.logger.Error("Error actualizando campaign_mkt", "err", err)
		return
	}
	p.logger.Info("campaign_mkt actualizado", "value", updated)
}

func (p *Pipedrive) createDeal(c *models.Communication, userId uint32, personId uint32) (*Deal, error) {
	title := c.Propiedad.Titulo.String
	if title == "" {
		title = "Trato con " + c.Nombre
	}

	payload := map[string]interface{}{
		"title":     title,
		"user_id":    userId,
		"person_id":  personId,
		"value":     c.Propiedad.Precio.String,
		"currency":  "MXN",
		"visible_to": "3", //Para todos
	}

	if mktCampaign, exists := findCampaignMkt(c.Message.String); exists {
		p.logger.Debug("MKT Campaign found on message", "campaign", mktCampaign)
		payload[customDealFields["campaign_mkt"]] = mktCampaign
	}

	var deal Deal
	err := p.makeRequest("POST", "deals", payload, &deal)

	if err != nil {
		return nil, err
	}
	return &deal, nil
}

// We want to reopens the deals in the first stage or a specific pipeline
func (p *Pipedrive) reopenDeal(c *models.Communication, deal *Deal) {
    // Get the first stage of the same pipeline of the deal
    stages, err := p.GetStages(StageFilter{
        PipelineId: deal.PipelineId,
        SortBy: "order_nr",
    })
    if err != nil || len(stages) == 0{
        p.logger.Error(fmt.Sprintf("cannot get the stages for pipeline (%d)", deal.PipelineId), "err", err.Error())
        return
    }
    p.logger.Debug(fmt.Sprintf("El deal está perdido, reabriendo en el stage %d", stages[0].ID))

	payload := map[string]interface{}{
		"status":   Open,
		"stage_id": stages[0].ID,
	}

	err = p.makeRequest("PUT", fmt.Sprintf("deals/%d/", deal.Id), payload, deal)
	if err != nil {
		p.logger.Error("Error reabriendo el deal", "err", err)
		return
	}
	p.logger.Info("Deal reabierto con exito")

	if mktCampaign, exists := findCampaignMkt(c.Message.String); exists {
		p.logger.Debug("MKT Campaign found on message", "campaign", mktCampaign)
		p.updateDealCampaignMkt(deal, mktCampaign)
	}
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
