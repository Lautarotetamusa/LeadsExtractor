package pipedrive

import "fmt"

type User struct{
    Id      uint32 `json:"id"`
    Name    string `json:"name"`
    Email   string `json:"email"`
}

func (p *Pipedrive) GetUserPersons(u *User) ([]Person, error) {
    var persons []Person
    url := fmt.Sprintf("persons?user_id=%d&limit=1000", u.Id)

    if err := p.makeRequest("GET", url, nil, &persons); err != nil {
        return nil, err
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

func (p *Pipedrive) GetUserByEmail(email string) (*User, error){
    var users []User
    url := fmt.Sprintf("users/find?term=%s&search_by_email=1", email)

    err := p.makeRequest("GET", url, nil, &users)

    if err != nil || len(users) == 0{
        return nil, fmt.Errorf("el usuario con email: %s, no existe", email)
    }

    return &users[0], nil
}
