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

type Pagination struct {
    Start   int `json:"start"`
    Limit	int `json:"limit"`
    MoreItemsInCollection bool  `json:"more_items_in_collection"`
    NextStart int `json:"next_start"` 
}

type AdditionalData struct {
    Pagination *Pagination `json:"pagination"`
}

type UsersPersonsResponse struct {
    Data []Person `json:"data"`
    AdditionalData *AdditionalData `json:"additional_data"`
}

func (p *Pipedrive) GetUserPersons(u *User) ([]Person, error) {
    var persons []Person
    moreItems := true
    start := 0

    for moreItems {
        fmt.Printf("Start = %d\n", start)
        url := baseUrl + fmt.Sprintf("persons?user_id=%d&start=%d", u.Id, start)

        p.refreshToken()
        req, err := http.NewRequest("GET", url, nil)
        if err != nil {
            return nil, err
        }

        req.Header.Add("Accept", "application/json")
        req.Header.Add("Content-Type", "application/json") 
        req.Header.Add("x-api-token", p.apiToken)

        res, err := p.client.Do(req)
        if err != nil {
            return nil, err
        }
        defer res.Body.Close()

        var response UsersPersonsResponse
        if err = json.NewDecoder(res.Body).Decode(&response); err != nil{
            return nil, err
        }

        moreItems = response.AdditionalData.Pagination.MoreItemsInCollection
        start = response.AdditionalData.Pagination.NextStart
        persons = append(persons, response.Data...)
    }

    return persons, nil
}

func (p *Pipedrive) GetUserDeals(u *User) ([]Deal, error) {
    var deals []Deal
    url := fmt.Sprintf("deals?user_id=%d&limit=1000", u.Id)

    if err := p.makeRequest("GET", url, nil, &deals); err != nil {
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
    fmt.Printf("|%s|\n", email);
    var users []User
    url := fmt.Sprintf("users/find?term=%s&search_by_email=1", email)

    err := p.makeRequest("GET", url, nil, &users)

    if err != nil || len(users) == 0{
        return nil, fmt.Errorf("el usuario con email: %s, no existe", email)
    }

    return &users[0], nil
}
