package pipedrive

import (
	"net/url"
	"strconv"
)

type Stage struct {
	ID                      int    `json:"id"`
	OrderNr                 int    `json:"order_nr"`
	Name                    string `json:"name"`
	ActiveFlag              bool   `json:"active_flag"`
	DealProbability         int    `json:"deal_probability"` // es un número, no un boolean
	PipelineID              int    `json:"pipeline_id"`
	RottenFlag              bool   `json:"rotten_flag"`
	RottenDays              int    `json:"rotten_days"`
	AddTime                 string `json:"add_time"`    // time.Time
	UpdateTime              string `json:"update_time"` // idem
	PipelineName            string `json:"pipeline_name"`
	PipelineDealProbability bool   `json:"pipeline_deal_probability"`
}

type SortDirection string
const (
    SortDirectionAsc  SortDirection = "asc"
    SortDirectionDesc SortDirection = "desc"
)

type StageFilter struct {
    PipelineId      int     `json:"pipeline_id"`
    SortBy          string  `json:"sort_by"`
    SortDirection   SortDirection  `json:"sort_direction"`
}

func (p *Pipedrive) GetStages(filter StageFilter) ([]*Stage, error) {
    params := url.Values{}
    if filter.PipelineId != 0 {
        params.Set("pipeline_id", strconv.Itoa(filter.PipelineId))
    }
    if filter.SortBy != "" {
        params.Set("sort_by", filter.SortBy)
    }
    if filter.SortDirection != "" {
        params.Set("sort_mode", string(filter.SortDirection))
    }

    endpoint := "stages"
    if encoded := params.Encode(); encoded != "" {
        endpoint += "?" + encoded
    }

    var stages []*Stage
    err := p.makeRequest("GET", endpoint, nil, &stages)
    if err != nil {
        return nil, err
    }

    return stages, nil
}
