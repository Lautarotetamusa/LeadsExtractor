package pipedrive

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct{
    Id      uint32 `json:"id"`
    Name    string `json:"name"`
    Email   string `json:"email"`
}

type PaginatedPersonResponse struct {
    Success bool                    `json:"success"`
    Data            []Person           `json:"data"`
    AdditionalData  *AdditionalData `json:"additional_data"`
}
func (p *Pipedrive) getPersonsPaginated(url string, persons *[]Person) error {
    moreItems := true
    start := 0

    for moreItems {
        p.logger.Debug("getting page", "page", start)
        url := baseUrl + fmt.Sprintf("%s&start=%d", url, start)

        p.refreshToken()
        req, err := http.NewRequest("GET", url, nil)
        if err != nil {
            return  err
        }

        req.Header.Add("Accept", "application/json")
        req.Header.Add("Content-Type", "application/json") 
        req.Header.Add("x-api-token", p.apiToken)

        res, err := p.client.Do(req)
        if err != nil {
            return err
        }
        defer res.Body.Close()

        var response PaginatedPersonResponse
        if err = json.NewDecoder(res.Body).Decode(&response); err != nil{
            return err
        }

        moreItems = response.AdditionalData.Pagination.MoreItemsInCollection
        start = response.AdditionalData.Pagination.NextStart

        *persons = append(*persons, response.Data...)
    }

    return nil
}
type PaginatedDealResponse struct {
    Success bool                    `json:"success"`
    Data            []Deal           `json:"data"`
    AdditionalData  *AdditionalData `json:"additional_data"`
}
func (p *Pipedrive) getDealsPaginated(url string, deals *[]Deal) error {
    moreItems := true
    start := 0

    for moreItems {
        p.logger.Debug("getting page", "page", start)
        url := baseUrl + fmt.Sprintf("%s&start=%d", url, start)

        p.refreshToken()
        req, err := http.NewRequest("GET", url, nil)
        if err != nil {
            return  err
        }

        req.Header.Add("Accept", "application/json")
        req.Header.Add("Content-Type", "application/json") 
        req.Header.Add("x-api-token", p.apiToken)

        res, err := p.client.Do(req)
        if err != nil {
            return err
        }
        defer res.Body.Close()

        var response PaginatedDealResponse
        if err = json.NewDecoder(res.Body).Decode(&response); err != nil{
            return err
        }

        moreItems = response.AdditionalData.Pagination.MoreItemsInCollection
        start = response.AdditionalData.Pagination.NextStart

        *deals = append(*deals, response.Data...)
    }

    return nil
}

func (p *Pipedrive) GetUserPersons(u *User) ([]Person, error) {
    var persons []Person
    url := fmt.Sprintf("persons?user_id=%d", u.Id)
    err := p.getPersonsPaginated(url, &persons) 
    if err != nil {
        return nil, err
    }
    return persons, nil
}

func (p *Pipedrive) GetUserDeals(u *User) ([]Deal, error) {
    var deals []Deal
    url := fmt.Sprintf("deals?user_id=%d", u.Id)
    err := p.getDealsPaginated(url, &deals) 
    if err != nil {
        return nil, err
    }
    return deals, nil
}

func (p *Pipedrive) GetUsers() ([]User, error){
    var users []User
    if err := p.makeRequest("GET", "users", nil, &users); err != nil {
        return nil, err
    }

    return users, nil
}

func (p *Pipedrive) GetUserByEmail(email string) (*User, error){
    var users []User
    url := fmt.Sprintf("users/find?term=%s&search_by_email=1", email)

    err := p.makeRequest("GET", url, nil, &users)

    if err != nil || len(users) == 0{
        return nil, fmt.Errorf("el usuario con email: %s, no existe", email)
    }

    return &users[0], nil
}
