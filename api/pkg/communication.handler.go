package pkg

import (
	"encoding/json"
	"leadsextractor/models"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/schema"
)

type QueryParam struct{
    DateFrom    time.Time   `schema:"fecha_from" db:"dateFrom"`
    DateTo      time.Time   `schema:"fecha_to" db:"dateTo"`
    AsesorPhone string      `schema:"asesor_phone" db:"asesorPhone"`
    AsesorName  string      `schema:"asesor_name" db:"asesorName"`
    Fuente      string      `schema:"fuente" db:"fuente"`
    Nombre      string      `schema:"nombre" db:"nombre"`
    Telefono    string      `schema:"telefono" db:"telefono"`
    IsNew       bool        `schema:"is_new" db:"isNew"`
    Page        int         `schema:"page" db:"page"`
}

type ListResponse = struct {
    Success     bool                    `json:"success"`
    Pagination  Pagination              `json:"pagination"`
    Data        []models.Communication  `json:"data"`
}

type Pagination struct{
    Page        int `json:"page"`
    PageSize    int `json:"page_size"`
    Items       int `json:"items"`
}

const pageSize = 20
const baseListQuery = ` 
SELECT 
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

// 2 de enero del 2006, se usa siempre esta fecha por algun motivo extraño xd
const timeLayout string = "2006-01-02"
var timeConverter = func(value string) reflect.Value {
    if v, err := time.Parse(timeLayout, value); err == nil {
        return reflect.ValueOf(v)
    }
    return reflect.Value{} // this is the same as the private const invalidType
}

func buildQuery(params QueryParam) (string, map[string]interface{}) {
	query := baseListQuery
	var whereClauses []string
	queryParams := make(map[string]interface{})

	if !params.DateFrom.IsZero() {
		whereClauses = append(whereClauses, "C.created_at >= :dateFrom")
		queryParams["dateFrom"] = params.DateFrom
	}
	if !params.DateTo.IsZero() {
		whereClauses = append(whereClauses, "C.created_at <= :dateTo")
		queryParams["dateTo"] = params.DateTo
	}
	if params.IsNew {
		whereClauses = append(whereClauses, "C.new_lead = :isNew")
		queryParams["isNew"] = params.IsNew
	}
	if params.Telefono != "" {
		whereClauses = append(whereClauses, "L.phone LIKE :telefono")
		queryParams["telefono"] = "%"+params.Telefono+"%"
	}
	if params.Fuente != "" {
		whereClauses = append(whereClauses, "IF(S.type = 'property', P.portal, S.type) LIKE :fuente")
		queryParams["fuente"] = "%"+params.Fuente+"%"
	}
	if params.AsesorPhone != "" {
		whereClauses = append(whereClauses, "A.phone LIKE :asesorPhone")
		queryParams["asesorPhone"] = "%"+params.AsesorPhone+"%"
	}
	if params.AsesorName != "" {
		whereClauses = append(whereClauses, "A.name LIKE :asesorName")
		queryParams["asesorName"] = "%"+params.AsesorName+"%"
	}
	if params.Nombre != "" {
		whereClauses = append(whereClauses, "L.name LIKE :nombre")
		queryParams["nombre"] = "%"+params.Nombre+"%"
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += " ORDER BY C.id DESC"
	if params.Page > 0 {
		offset := (params.Page - 1) * pageSize
		query += " LIMIT :pageSize OFFSET :offset"
		queryParams["pageSize"] = pageSize
		queryParams["offset"] = offset
	}else{
		query += " LIMIT :pageSize"
		queryParams["pageSize"] = pageSize
    }

	return query, queryParams
}

func (s *Server) GetCommunications(w http.ResponseWriter, r *http.Request) error {
    var (
        params QueryParam
        decoder = schema.NewDecoder()
    )
    //Lo hago asi para que si no encuentra nada devuelva []
    communications := make([]models.Communication, 0)

    decoder.RegisterConverter(time.Time{}, timeConverter)
    err := decoder.Decode(&params, r.URL.Query())
    if (err != nil){
        return err;
    }

    query, queryParams := buildQuery(params)

    rows, err := s.Store.Db.NamedQuery(query, queryParams);
    if (err != nil){
        return err;
    }
    defer rows.Close()

	for rows.Next() {
		var comm models.Communication
		err := rows.StructScan(&comm)
		if err != nil {
			return err
		}
		communications = append(communications, comm)
	}

    w.Header().Set("Content-Type", "application/json")
    res := ListResponse{
        Success: true,
        Pagination: Pagination{
            Page: params.Page,
            PageSize: pageSize,
            Items: len(communications),
        },
        Data: communications,
    }
    json.NewEncoder(w).Encode(res)
	return nil
}

func (s *Server) NewCommunication(c *models.Communication) error {
	source, err := s.Store.GetSource(c)
	if err != nil {
		return err
	}

	lead, err := s.Store.InsertOrGetLead(s.roundRobin, c)
	if err != nil {
		return err
	}

	c.Asesor = lead.Asesor
    c.Cotizacion = lead.Cotizacion
    c.Email = lead.Email
    c.Nombre = lead.Name

    go s.flowHandler.manager.RunMainFlow(c)
        
    if err = s.Store.InsertCommunication(c, source); err != nil {
        s.logger.Error(err.Error(), "path", "InsertCommunication")
        return err
    }
    return nil
}

func (s *Server) HandleNewCommunication(w http.ResponseWriter, r *http.Request) error {
	c := &models.Communication{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(c); err != nil {
		return err
	}
    
    if err := s.NewCommunication(c); err != nil {
        return err
    }

    successResponse(w, "communication created succesfuly", c)
	return nil
}
