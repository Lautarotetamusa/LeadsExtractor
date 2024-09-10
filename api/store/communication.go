package store

import (
	"leadsextractor/models"
	"strings"
	"time"
)

type QueryParam struct{
    DateFrom    time.Time   `schema:"fecha_from" json:"fecha_from,omitempty" db:"dateFrom"`
    DateTo      time.Time   `schema:"fecha_to" json:"fecha_to,omitempty" db:"dateTo"`
    AsesorPhone string      `schema:"asesor_phone" json:"asesor_phone,omitempty" db:"asesorPhone"`
    AsesorName  string      `schema:"asesor_name" json:"asesor_name,omitempty" db:"asesorName"`
    Fuente      string      `schema:"fuente" json:"fuente,omitempty" db:"fuente"`
    Nombre      string      `schema:"nombre" json:"nombre,omitempty" db:"nombre"`
    Telefono    string      `schema:"telefono" json:"telefono,omitempty" db:"telefono"`
    IsNew       *bool       `schema:"is_new" json:"is_new,omitempty" db:"isNew"`
    Page        int         `schema:"page" json:"page,omitempty" db:"page"`
    PageSize    int         `schema:"page_size" json:"page_size,omitempty"`
}

type Query struct {
    query   string
    params  map[string]interface{}
}

const defaultPageSize = 10
const minPageSize = 5
const maxPageSize = 100
//DATE_SUB(C.lead_date, INTERVAL 6 HOUR) as "lead_date", // falta este campo, hay muchos que es NULL
const selectQuery = ` 
SELECT 
    C.id,
    DATE_SUB(C.created_at, INTERVAL 6 HOUR) as "created_at",
    C.new_lead,
    L.cotizacion,
    A.name as "asesor.name", 
    A.phone as "asesor.phone", 
    A.email as "asesor.email",
    utm_source as "utm.utm_source",
    utm_medium as "utm.utm_medium",
    utm_campaign as "utm.utm_campaign",
    IF(S.type = "property", P.portal, S.type) as "fuente",
    L.name, 
    C.url, 
    L.phone,
    L.email,
    IFNULL(P.portal_id, "") as "propiedad.portal_id", 
    IFNULL(P.title, "") as "propiedad.title", 
    IFNULL(P.price, "") as "propiedad.price", 
    IFNULL(P.ubication, "") as "propiedad.ubication", 
    IFNULL(P.url, "") as "propiedad.url", 
    IFNULL(P.tipo, "") as "propiedad.tipo",
    C.zones as "busquedas.zones", 
    C.mt2_terrain as "busquedas.mt2_terrain", 
    C.mt2_builded as "busquedas.mt2_builded", 
    C.baths as "busquedas.baths", 
    C.rooms as "busquedas.rooms"
`;
const countQuery = "SELECT COUNT(*)";
const joinQuery = `
FROM Communication C
INNER JOIN Leads L
    ON C.lead_phone = L.phone
INNER JOIN Source S
    ON C.source_id = S.id
INNER JOIN Asesor A
    ON L.asesor = A.phone
LEFT JOIN Property P
    ON S.property_id = P.id
`;

func NewQuery(baseQuery string) Query {
    return Query {
        query: baseQuery,
        params: map[string]interface{}{},
    }
}

// Validate that the communication matches the params
func (p *QueryParam) Matches(c *models.Communication) bool {
    return  (p.IsNew == nil     || *p.IsNew == c.IsNew) && 
            (p.Telefono == ""   ||  p.Telefono == c.Telefono.String()) && 
            (p.Fuente == ""     ||  p.Fuente == c.Fuente) && 
            (p.AsesorPhone == "" || p.AsesorPhone == c.Asesor.Phone.String()) && 
            (p.AsesorName == "" ||  p.AsesorName == c.Asesor.Name) &&  
            (p.Nombre == ""     ||  p.Nombre == c.Nombre)
}

func (q *Query) buildWhere(params *QueryParam) {
    var whereClauses []string

	if !params.DateFrom.IsZero() {
		whereClauses = append(whereClauses, "DATE_SUB(C.created_at, INTERVAL 6 HOUR) >= :dateFrom")
		q.params["dateFrom"] = params.DateFrom
	}
	if !params.DateTo.IsZero() {
		whereClauses = append(whereClauses, "DATE_SUB(C.created_at, INTERVAL 6 HOUR) <= :dateTo")
		q.params["dateTo"] = params.DateTo
	}
	if params.IsNew != nil{
		whereClauses = append(whereClauses, "C.new_lead = :isNew")
		q.params["isNew"] = params.IsNew
	}
	if params.Telefono != "" {
		whereClauses = append(whereClauses, "L.phone LIKE :telefono")
		q.params["telefono"] = "%"+params.Telefono+"%"
	}
	if params.Fuente != "" {
		whereClauses = append(whereClauses, "IF(S.type = 'property', P.portal, S.type) LIKE :fuente")
		q.params["fuente"] = "%"+params.Fuente+"%"
	}
	if params.AsesorPhone != "" {
		whereClauses = append(whereClauses, "A.phone LIKE :asesorPhone")
		q.params["asesorPhone"] = "%"+params.AsesorPhone+"%"
	}
	if params.AsesorName != "" {
		whereClauses = append(whereClauses, "A.name LIKE :asesorName")
		q.params["asesorName"] = "%"+params.AsesorName+"%"
	}
	if params.Nombre != "" {
		whereClauses = append(whereClauses, "L.name LIKE :nombre")
		q.params["nombre"] = "%"+params.Nombre+"%"
	}

	if len(whereClauses) > 0 {
		q.query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
}

func (q *Query) buildPagination(params *QueryParam) {
    if params.Page == 0 {
        params.Page = 1
    }

	q.query += " ORDER BY C.id DESC"
    if (params.PageSize < minPageSize || params.PageSize > maxPageSize){
        params.PageSize = defaultPageSize;
    }

    offset := (params.Page - 1) * params.PageSize
    q.query += " LIMIT :pageSize OFFSET :offset"
    q.params["pageSize"] = params.PageSize
    q.params["offset"] = offset
}

func (s *Store) InsertCommunication(c *models.Communication, source *models.Source) error {
	query := `INSERT INTO Communication(lead_phone, source_id, new_lead, lead_date, utm_source, utm_medium, utm_campaign, url, zones, mt2_terrain, mt2_builded, baths, rooms) 
    VALUES (:lead_phone, :source_id, :new_lead, :lead_date, :utm_source, :utm_medium, :utm_campaign, :url, :zones, :mt2_terrain, :mt2_builded, :baths, :rooms)`
    _, err := s.db.NamedExec(query, map[string]interface{}{
		"lead_phone":  c.Telefono,
		"source_id":   source.Id,
		"new_lead":    c.IsNew,
        "lead_date":   c.FechaLead,  
        "utm_source":  c.Utm.Source,
        "utm_medium":  c.Utm.Medium,
        "utm_campaign":c.Utm.Campaign,
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
    s.logger.Info("communication saved")
    return nil
}

func (s *Store) GetCommunications(params *QueryParam) ([]models.Communication, error) {
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

func (s *Store) GetCommunicationsCount(params *QueryParam) (int, error) {
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

func (s *Store) ExistsCommunication(params *QueryParam) bool {
    count, err := s.GetCommunicationsCount(params)
    if err != nil {
        return false
    }
    return count > 0
}
