package store

import (
	"leadsextractor/models"
	"strings"
	"time"
)

type QueryParam struct{
    DateFrom    time.Time   `schema:"fecha_from" db:"dateFrom"`
    DateTo      time.Time   `schema:"fecha_to" db:"dateTo"`
    AsesorPhone string      `schema:"asesor_phone" db:"asesorPhone"`
    AsesorName  string      `schema:"asesor_name" db:"asesorName"`
    Fuente      string      `schema:"fuente" db:"fuente"`
    Nombre      string      `schema:"nombre" db:"nombre"`
    Telefono    string      `schema:"telefono" db:"telefono"`
    IsNew       *bool       `schema:"is_new" db:"isNew"`
    Page        int         `schema:"page" db:"page"`
    PageSize    int         `schema:"page_size"`
}

type Query struct {
    query   string
    params  map[string]interface{}
}

const defaultPageSize = 10
const minPageSize = 5
const maxPageSize = 100
const selectQuery = ` 
SELECT 
    C.id,
    C.created_at, 
    C.new_lead,
    L.cotizacion,
    A.name as "asesor.name", 
    A.phone as "asesor.phone", 
    A.email as "asesor.email",
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

func (q *Query) buildWhere(params *QueryParam) {
    var whereClauses []string

	if !params.DateFrom.IsZero() {
		whereClauses = append(whereClauses, "C.created_at >= :dateFrom")
		q.params["dateFrom"] = params.DateFrom
	}
	if !params.DateTo.IsZero() {
		whereClauses = append(whereClauses, "C.created_at <= :dateTo")
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
	query := `INSERT INTO Communication(lead_phone, source_id, new_lead, lead_date, url, zones, mt2_terrain, mt2_builded, baths, rooms) 
    VALUES (:lead_phone, :source_id, :new_lead, :lead_date, :url, :zones, :mt2_terrain, :mt2_builded, :baths, :rooms)`
    _, err := s.db.NamedExec(query, map[string]interface{}{
		"lead_phone":  c.Telefono,
		"source_id":   source.Id,
		"new_lead":    c.IsNew,
		"lead_date":   c.FechaLead,
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

