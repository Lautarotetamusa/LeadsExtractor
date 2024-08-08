package pkg

import "net/http"

func (s *Server) HandleGetNewLeadsMetric(w http.ResponseWriter, r *http.Request) error {
	metric, err := s.Store.GetNewLeads()
	if err != nil {
		return err
	}

	dataResponse(w, metric)
	return nil
}
