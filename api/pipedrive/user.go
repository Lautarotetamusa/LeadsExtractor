package pipedrive

import "fmt"

type User struct{
    Id      uint32 `json:"id"`
    Name    string `json:"name"`
    Email   string `json:"email"`
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
