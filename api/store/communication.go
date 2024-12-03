package store

import (
	"fmt"
	"leadsextractor/models"
	"reflect"
	"strings"
	"time"
)

type QueryParam struct {
    DateFrom    time.Time `schema:"fecha_from" json:"fecha_from,omitempty" db:"dateFrom" select:"C.created_at"`
    DateTo      time.Time `schema:"fecha_to" json:"fecha_to,omitempty" db:"dateTo" select:"C.created_at"`
    AsesorPhone string    `schema:"asesor_phone" json:"asesor_phone,omitempty" db:"asesorPhone" select:"A.phone"`
    AsesorName  string    `schema:"asesor_name" json:"asesor_name,omitempty" db:"asesorName" select:"A.name"`
    Fuente      string    `schema:"fuente" json:"fuente,omitempty" db:"fuente" select:"IF(S.type = 'property', P.portal, S.type)"`
    Nombre      string    `schema:"nombre" json:"nombre,omitempty" db:"nombre" select:"L.name"`
    Telefono    string    `schema:"telefono" json:"telefono,omitempty" db:"telefono" select:"L.phone"`
    IsNew       *bool     `schema:"is_new" json:"is_new,omitempty" db:"isNew" select:"C.new_lead"`
    Page        int       `schema:"page" json:"page,omitempty" db:"page"`
    PageSize    int       `schema:"page_size" json:"page_size,omitempty"`
    Message     string    `schema:"message" json:"message" select:"M.text"`
    UtmSource   string    `schema:"utm_source" json:"utm_source,omitempty" db:"utm_source" select:"C.utm_source"`
    UtmMedium   string    `schema:"utm_medium" json:"utm_medium,omitempty" db:"utm_medium" select:"C.utm_medium"`
    UtmCampaign string    `schema:"utm_campaign" json:"utm_campaign,omitempty" db:"utm_campaign" select:"C.utm_campaign"`
    UtmAd       string    `schema:"utm_ad" json:"utm_ad,omitempty" db:"utm_ad" select:"C.utm_ad"`
    UtmChannel  string    `schema:"utm_channel" json:"utm_channel,omitempty" db:"utm_channel" select:"C.utm_channel"`
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
    utm_channel as "utm.utm_channel",
    utm_ad as "utm.utm_ad",
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
    C.rooms as "busquedas.rooms",
    M.text as "message"
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
LEFT Join Message M
    ON M.id_communication = C.id
`;

const groupByLeadPhone = " GROUP BY L.phone"

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
            (p.Nombre == ""     ||  p.Nombre == c.Nombre) &&
            (p.UtmCampaign == "" || p.UtmCampaign == c.Utm.Campaign.String) &&
            (p.UtmAd == "" || p.UtmAd == c.Utm.Ad.String) &&
            (p.UtmChannel == "" || p.UtmChannel == c.Utm.Channel.String) &&
            (p.UtmMedium == "" || p.UtmMedium == c.Utm.Medium.String) &&
            (p.UtmSource == "" || p.UtmSource == c.Utm.Source.String) &&
            (p.Message == "" || p.Message == c.Message.String)
}

func encloseWords(input string) string {
	words := strings.Split(input, ",")
	
	for i, word := range words {
		words[i] = `"` + strings.TrimSpace(word) + `"`
	}

	return strings.Join(words, ",")
}

func (q *Query) buildWhere(params *QueryParam) {
    var whereClauses []string

    t := reflect.TypeOf(*params)
    paramValue := reflect.ValueOf(*params)

    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        fieldValue := paramValue.Field(i)
        dbTag := field.Tag.Get("db")
        selectTag := field.Tag.Get("select")

        if dbTag == "" || selectTag == "" {
            continue
        }

        switch fieldValue.Kind(){
        case reflect.String:
            if fieldValue.String() == "" {
                continue
            }
            values := strings.Split(fieldValue.String(), ",")
            var queryWhere []string
            for i, value := range values {
                tag := fmt.Sprintf("%s%v", dbTag, i)
                queryWhere = append(queryWhere, fmt.Sprintf("%s LIKE :%s%v", selectTag, dbTag, i))
                q.params[tag] = "%"+value+"%"
            }

            whereClauses = append(whereClauses, "("+strings.Join(queryWhere, " OR ")+")")
        case reflect.Ptr:
            if boolPtr, ok := fieldValue.Interface().(*bool); ok && boolPtr != nil {
                whereClauses = append(whereClauses, selectTag+" = :"+dbTag)
                q.params[dbTag] = *boolPtr
            }
        case reflect.Struct:
            if field.Type == reflect.TypeOf(time.Time{}) && !fieldValue.Interface().(time.Time).IsZero() {
                var comp string
                if field.Name == "DateFrom" {
                    comp = ">="
                } else if field.Name == "DateTo" {
                    comp = "<="
                }
                queryStr := fmt.Sprintf("DATE_SUB(%s, INTERVAL 6 HOUR) %s :%s",selectTag, comp, dbTag)
                whereClauses = append(whereClauses, queryStr)
                q.params[dbTag] = fieldValue.Interface()
            }
        }
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
    fields := []string{
        "lead_phone", "source_id", "new_lead", "lead_date", "utm_source", 
        "utm_medium", "utm_campaign", "utm_ad", "utm_channel",
        "url", "zones", "mt2_terrain", "mt2_builded", "baths", "rooms",
    }

    query := fmt.Sprintf(
        "INSERT INTO Communication (%s) VALUES (:%s)", 
        strings.Join(fields, ", "), strings.Join(fields, ", :"),
    )

	res, err := s.db.NamedExec(query, map[string]interface{}{
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

func (s *Store) GetAllCommunications(params *QueryParam) ([]models.Communication, error) {
    //Lo hago asi para que si no encuentra nada devuelva []
    communications := make([]models.Communication, 0)

    query := NewQuery(selectQuery + joinQuery)
    query.buildWhere(params)

    stmt, err := s.db.PrepareNamed(query.query)
    if err != nil {
        return nil, err
    }
    if err := stmt.Select(&communications, query.params); err != nil {
        return nil, err
    }

    return communications, nil
}

func (s *Store) GetAllDistinctCommunications(params *QueryParam) ([]models.Communication, error) {
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
