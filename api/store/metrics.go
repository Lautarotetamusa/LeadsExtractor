package store

import (
	"time"
)

type Value struct {
    At      string `db:"at" json:"at"`
    Cant    uint32 `db:"cant" json:"cant"`
}

type Metric struct {
    Name        string          `json:"name"`
    Total       uint32          `json:"total"`
    TimeSpan    time.Duration   `json:"time_span"`
    Values      []Value         `json:"values"`
}

const newLeadsQuery = `
SELECT 
    DATE(C.created_at) AS at,
    COUNT(C.id) AS cant
FROM 
    Communication C
WHERE 
    C.new_lead = TRUE
GROUP BY 
    DATE(C.created_at)
ORDER BY 
    DATE(C.created_at);
`;

func (s *Store) GetNewLeads() (*Metric, error) {
    values := []Value{}
	if err := s.db.Select(&values, newLeadsQuery); err != nil {
		return nil, err
	}

    var count uint32
    query := `SELECT COUNT(*) FROM Communication WHERE new_lead = true` 
	if err := s.db.QueryRow(query).Scan(&count); err != nil {
		return nil, err
	}

    metric := Metric{
        Name: "new_leads",
        Total: count, 
        Values: values,
        // TimeSpan: time.Duration(1 day),
    }
	return &metric, nil
}
