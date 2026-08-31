package logs

import (
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

type LogEntry struct {
	Msg    string    `bson:"msg" json:"msg"`
	Time   time.Time `bson:"time" json:"time"`
	Level  string    `bson:"level" json:"level"`
	Module string    `bson:"module" json:"module"`
}

type Pagination struct {
	ItemsCount int   `json:"items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	IsLastPage bool  `json:"is_last_page"`
}

type LogsHandler struct {
	collection *mongo.Collection
	context    context.Context
}

func NewLogsHandler(ctx context.Context, c *mongo.Collection) *LogsHandler {
	return &LogsHandler{
		collection: c,
		context:    ctx,
	}
}

func (p *Pagination) setItemsCount(count int) {
	p.ItemsCount = count
	p.IsLastPage = p.ItemsCount != p.PageSize
}

func buildPagination(opts *options.FindOptions, query url.Values) *Pagination {
	limit, err := strconv.Atoi(query.Get("page_size"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	skip := (page - 1) * limit
	opts.SetSkip(int64(skip)).SetLimit(int64(limit))

	return &Pagination{
		Page:     page,
		PageSize: limit,
	}
}

// TODO: Parsear correctamente las fechas
func buildFilter(query url.Values) bson.M {
	filter := bson.M{}

	for field, values := range query {
		if field == "page" || field == "page_size" {
			continue
		}

		// Si la request es GET /?name=pedro&name=juan
		// query[name] = {pedro, juan}
		if len(values) == 1 {
			// GET /?time_gt=2024-12-01 ==> {time: {"$gt": 2024-12-01}}
			query_strs := strings.Split(field, "_")
			if len(query_strs) == 1 {
				filter[field] = values[0]
			} else if slices.Contains([]string{"gt", "lt"}, query_strs[1]) {
				// Si el filtro[key] YA es un objeto bson.M
				if f, ok := filter[query_strs[0]].(bson.M); ok {
					f["$"+query_strs[1]] = values[0] // Lo agregamos
				} else {
					filter[query_strs[0]] = bson.M{"$" + query_strs[1]: values[0]}
				}
			}
		} else {
			filter[field] = bson.M{"$in": values}
		}
	}
	return filter
}

func (h *LogsHandler) GetLogs(w http.ResponseWriter, r *http.Request) error {
	opts := options.Find().SetSort(bson.M{"_id": -1}) // Descendente

	query := r.URL.Query()

	filter := buildFilter(query)

	pagination := buildPagination(opts, query)

	cur, err := h.collection.Find(h.context, filter, opts)
	defer cur.Close(h.context)
	if err != nil {
		return err
	}

	total, err := h.collection.CountDocuments(h.context, filter)
	if err != nil {
		return err
	}
	pagination.Total = total

	logs := []bson.M{} // Para no devolver null si no existen logs
	if err = cur.All(h.context, &logs); err != nil {
		return err
	}
	pagination.setItemsCount(len(logs))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"data":       logs,
		"pagination": pagination,
		"success":    true,
	})
	return nil
}
