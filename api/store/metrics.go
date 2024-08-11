package store

import (
	"fmt"
	"time"
)

type Value struct {
    At      string `db:"at" json:"at"`
    Cant    uint32 `db:"cant" json:"cant"`
}

type Metric struct {
    Total       uint32          `json:"total"`
    TimeSpan    time.Duration   `json:"time_span"`
    Values      []Value         `json:"values"`
}

const baseSelect = `
SELECT 
    DATE(C.created_at) AS at,
    COUNT(C.id) AS cant
FROM 
    Communication C`;

const baseCount = "SELECT COUNT(*) FROM Communication C"

const baseGroup = `
GROUP BY 
    DATE(C.created_at)
ORDER BY 
    DATE(C.created_at);`;

func getMetricValue(s *Store, params *QueryParam) ([]Value, error) {
    values := []Value{}
    q := NewQuery(baseSelect)
    q.buildWhere(params)
    q.query += baseGroup

    stmt, err := s.db.PrepareNamed(q.query)
    if err != nil {
        return nil, err
    }
    if err := stmt.Select(&values, q.params); err != nil {
        return nil, err
    }
    return values, nil
}

func getMetricCount(s *Store, params *QueryParam) (*uint32, error) {
    q := NewQuery(baseCount)
    q.buildWhere(params)
    fmt.Println()
    fmt.Println(q.query)

    stmt, err := s.db.PrepareNamed(q.query)
    var count uint32
    if err != nil {
        return nil, err
    }
    if err := stmt.QueryRow(q.params).Scan(&count); err != nil {
        return nil, err
    }
    return &count, nil
}

func (s *Store) GetMetrics(params *QueryParam) (*Metric, error) {
    values, err := getMetricValue(s, params)
    if err != nil {
        return nil, err
    }
    count, err := getMetricCount(s, params)
    if err != nil {
        return nil, err
    }
    metric := Metric{
        Total: *count, 
        Values: values,
        // TimeSpan: time.Duration(1 day),
    }
    return &metric, nil
}
