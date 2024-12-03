package pipedrive

import (
	"encoding/json"
	"fmt"
)

type DealStatus int

const (
    Open DealStatus = iota
    Won
    Lost
    Deleted
)
var (
    dealStatusName = map[DealStatus]string{
        Open: "open",
        Won: "won",
        Lost: "lost",
        Deleted: "deleted",
    }
    dealStatusValue = map[string]DealStatus{
        "open" : 0,
        "won": 1,
        "lost": 2,
        "delted": 3,
    }
)

func ParseDealStatus(s string) (DealStatus, error) {
	value, ok := dealStatusValue[s]
	if !ok {
		return DealStatus(0), fmt.Errorf("%q is not a valid card suit", s)
	}
	return DealStatus(value), nil
}

func (s DealStatus) String() string {
	return dealStatusName[s]
}

func (s *DealStatus) UnmarshalJSON(data []byte) (err error) {
	var status string
	if err := json.Unmarshal(data, &status); err != nil {
		return err
	}
	if *s, err = ParseDealStatus(status); err != nil {
		return err
	}
	return nil
}

func (s DealStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
