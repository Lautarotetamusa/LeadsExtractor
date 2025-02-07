package store

import (
	"fmt"
	"leadsextractor/models"
	"strings"

	"github.com/jmoiron/sqlx"
)

type CommunicationStorer interface {
    Insert(c *models.Communication, source *models.Source) error
    // Get all with pagination
    GetAll(params *QueryParam) ([]models.Communication, error)
    GetDistinct(params *QueryParam) ([]models.Communication, error)
    Count(params *QueryParam) (int, error)
    Exists(params *QueryParam) bool
}

type CommunicationStore struct {
    db *sqlx.DB
}

func NewCommStore(db *sqlx.DB) *CommunicationStore {
    return &CommunicationStore{
        db: db,
    }
}

var fields = []string{
    "lead_phone", "source_id", "new_lead", "lead_date", "utm_source", 
    "utm_medium", "utm_campaign", "utm_ad", "utm_channel",
    "url", "zones", "mt2_terrain", "mt2_builded", "baths", "rooms",
}

var insertQuery string = fmt.Sprintf(
    "INSERT INTO Communication (%s) VALUES (:%s)", 
    strings.Join(fields, ", "), strings.Join(fields, ", :"),
)

func (s *CommunicationStore) Insert(c *models.Communication, source *models.Source) error {
	res, err := s.db.NamedExec(insertQuery, map[string]interface{}{
		"lead_phone":  c.Telefono,
		"source_id":   source.Id,
		"new_lead":    c.IsNew,
        "lead_date":   c.FechaLead,  
        "utm_source":  c.Utm.Source,
        "utm_medium":  c.Utm.Medium,
        "utm_campaign":c.Utm.Campaign,
        "utm_ad":      c.Utm.Ad,
        "utm_channel": c.Utm.Channel,
		"url":         c.Link,
		"zones":       c.Busquedas.Zonas,
		"mt2_terrain": c.Busquedas.TotalArea,
		"mt2_builded": c.Busquedas.CoveredArea,
		"baths":       c.Busquedas.Banios,
		"rooms":       c.Busquedas.Recamaras,
	})
	if err != nil {
		return err
	}
    if id, err := res.LastInsertId(); err != nil{
        return err
    }else{
        c.Id = int(id)
    }

    return nil
}

func (s *CommunicationStore) GetAll(params *QueryParam) ([]models.Communication, error) {
    //Lo hago asi para que si no encuentra nada devuelva []
    communications := make([]models.Communication, 0)

    query := NewQuery(selectQuery + joinQuery)
    query.buildWhere(params)
    query.buildPagination(params)

    stmt, err := s.db.PrepareNamed(query.query)
    if err != nil {
        return nil, err
    }
    if err := stmt.Select(&communications, query.params); err != nil {
        return nil, err
    }

    return communications, nil
}

func (s *CommunicationStore) GetDistinct(params *QueryParam) ([]models.Communication, error) {
    //Lo hago asi para que si no encuentra nada devuelva []
    communications := make([]models.Communication, 0)

    query := NewQuery(selectQuery + joinQuery)
    query.buildWhere(params)
    query.query += groupByLeadPhone

    stmt, err := s.db.PrepareNamed(query.query)
    if err != nil {
        return nil, err
    }
    if err := stmt.Select(&communications, query.params); err != nil {
        return nil, err
    }

    return communications, nil
}

func (s *CommunicationStore) Count(params *QueryParam) (int, error) {
    var count int
    query := NewQuery(countQuery + joinQuery)
    query.buildWhere(params)

    stmt, err := s.db.PrepareNamed(query.query);
    if err != nil {
        return 0, err
    }
    if err := stmt.Get(&count, query.params); err != nil {
        return 0, err
    }

    return count, nil
}

func (s *CommunicationStore) Exists(params *QueryParam) bool {
    count, err := s.Count(params)
    if err != nil {
        return false
    }
    return count > 0
}
