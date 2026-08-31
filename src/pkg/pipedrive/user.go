package pipedrive

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct {
	Id    uint32 `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type paginatedResponse[T any] struct {
	Success        bool            `json:"success"`
	Data           []T             `json:"data"`
	AdditionalData *AdditionalData `json:"additional_data"`
}

func getPaginated[T any](p *Pipedrive, path string, items *[]T) error {
	moreItems := true
	start := 0

	for moreItems {
		p.logger.Debug("getting page", "page", start)
		url := baseUrl + fmt.Sprintf("%s&start=%d", path, start)

		p.refreshToken()
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}

		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("x-api-token", p.apiToken)

		res, err := p.client.Do(req)
		if err != nil {
			return err
		}

		var response paginatedResponse[T]
		err = json.NewDecoder(res.Body).Decode(&response)
		res.Body.Close()
		if err != nil {
			return err
		}

		moreItems = response.AdditionalData.Pagination.MoreItemsInCollection
		start = response.AdditionalData.Pagination.NextStart

		*items = append(*items, response.Data...)
	}

	return nil
}

func (p *Pipedrive) GetUserPersons(u *User) ([]Person, error) {
	var persons []Person
	if err := getPaginated(p, fmt.Sprintf("persons?user_id=%d", u.Id), &persons); err != nil {
		return nil, err
	}
	return persons, nil
}

func (p *Pipedrive) GetUserDeals(u *User) ([]Deal, error) {
	var deals []Deal
	if err := getPaginated(p, fmt.Sprintf("deals?user_id=%d", u.Id), &deals); err != nil {
		return nil, err
	}
	return deals, nil
}

func (p *Pipedrive) GetUsers() ([]User, error) {
	var users []User
	if err := p.makeRequest("GET", "users", nil, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (p *Pipedrive) GetUserByEmail(email string) (*User, error) {
	var users []User
	url := fmt.Sprintf("users/find?term=%s&search_by_email=1", email)

	if err := p.makeRequest("GET", url, nil, &users); err != nil || len(users) == 0 {
		return nil, fmt.Errorf("el usuario con email: %s, no existe", email)
	}

	return &users[0], nil
}
