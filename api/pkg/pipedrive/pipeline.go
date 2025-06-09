package pipedrive

type Pipeline struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	URLTitle       string `json:"url_title"`
	OrderNr        int    `json:"order_nr"`
	Active         bool   `json:"active"`
	DealProbability bool  `json:"deal_probability"`
	AddTime        string `json:"add_time"`    // time.Time 
	UpdateTime     string `json:"update_time"` // same
	Selected       bool   `json:"selected"`
}

func (p *Pipedrive) GetPipelines() ([]*Pipeline, error) {
    var pipelines []*Pipeline
    err := p.makeRequest("GET", "pipelines", nil, &pipelines)
    if err != nil {
        return nil, err
    }

    return pipelines, err
}
