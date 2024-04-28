package pipedrive

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"leadsextractor/models"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
    "io"
	"strconv"
	"strings"
	"time"
)

const baseUrl = "https://api.pipedrive.com/v1/"

type Pipedrive struct{
    Token       *Token 
    Auth        Auth
    ApiToken    string
    RedirectUri string
    Client      *http.Client
    State       string
}

type Token struct{
    AccessToken     string `json:"access_token,omitempty"`
    RefreshToken    string `json:"refresh_token,omitempty"`
    Scope           string `json:"scope,omitempty"`
    ExpiresIn       int64  `json:"expires_in,omitempty"` //Tiempo Unix() en el que expira el token
    ApiDomain       string `json:"api_domain,omitempty"`
}

type Auth struct{
    ClientId        string
    ClientSecret    string
    Value           string
}

type Response struct{
    Success bool        `json:"success"`
    Data    interface{} `json:"data"`
}

type User struct{
    Id      uint32 `json:"id"`
    Name    string `json:"name"`
    Email   string `json:"email"`
}

type Person struct{
    Id      uint32 `json:"id"`
    Name    string `json:"name"`
    Phone   []struct{
        Value   string  `json:"value"`
        Primary bool    `json:"primary"`
    } `json:"phone"`
    Email   []struct{
        Value   string  `json:"value"`
        Primary bool    `json:"primary"`
    } `json:"email"`
}

type FieldOption struct{
    Id      uint32 `json:"id"`
    Label   string `json:"label"`
}

var customFields = map[string]string{
    "fuente": "0a551c8c09663ce803d924f036af12c3cc6b8b73",
    "fecha lead": "a7b3035eaea7ae5cb3aeab97f8f91748aa8e427b",
    "link": "180b604d295a05730ea6f453a384a3dc78bb108c",
    "zona": "41a4e5fc09fbc6243074f02d9dc284b8f9b1a505",
    "ubicacion": "8274fd2ca686df722bb1e78c5479acaba1067058",
    "precio": "80b2cb3150d88f04e61e442b76d942e636596274",
};

var fuenteOptions = map[string]uint32{
    "vivanuncios": 21,
    "rebora": 22,
    "inmuebles24": 23,
    "lamudi": 24,
    "casasyterrenos": 25,
    "propiedades": 26,
    "whatsapp": 74,
    "ivr": 75,
};

func NewPipedrive(clientId string, clientSecret string, apiToken string, redirectUri string) *Pipedrive{
    client := fmt.Sprintf("%s:%s", clientId, clientSecret) 
    p := Pipedrive{
        Auth: Auth{
            ClientId: clientId,
            ClientSecret: clientSecret,
            Value: fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(client))),
        },
        Client: &http.Client{
            Timeout: 15 * time.Second,
        },
        RedirectUri: redirectUri,
        ApiToken: apiToken,
        Token: nil,
    } 
    p.loadToken()

    if p.Token != nil{
        p.refreshToken()        
    }
    return &p
}

func (p *Pipedrive) SaveCommunication(c *models.Communication){
    asesor, err := p.getUserByName(c.Asesor.Name)
    if err != nil{
        log.Printf("Error obteniendo el asesor %s", err)
        return
    }
    fmt.Printf("Asesor: %v\n", asesor)

    person, err := p.getPersonByNumber(c.Telefono)
    if err != nil{
        log.Println(err)
        log.Println("Creando persona en PipeDrive")
        person, err = p.createPerson(c, asesor.Id)
        if err != nil{
            log.Printf("Error creando persona %s", err)
            return
        }
    }
    fmt.Printf("Person: %v\n", person)

    log.Println("Cargando deal en PipeDrive")
    deal, err := p.createDeal(c, asesor.Id, person.Id)
    if err != nil{
        log.Printf("Error creando deal %s", err)
        return
    }
    fmt.Printf("Deal: %v\n", deal)

    log.Printf("Deal cargando correctamente en PipeDrive")
}

func (p *Pipedrive) saveToken(){
    //Hacemos que el campo expiresIn sea el tiempo Unix en el que expirará
    p.Token.ExpiresIn = time.Now().Unix() + int64(p.Token.ExpiresIn)
    fmt.Printf("%v\n", p.Token)

    file, _ := json.MarshalIndent(p.Token, "", "\t")
    fileName := fmt.Sprintf("%s.json", p.Auth.ClientId)

    err := ioutil.WriteFile(fileName, file, 0644)
    if err != nil{
        log.Fatal("No se pudo guardar el token en el archivo")
    }
}

func (p *Pipedrive) loadToken() *Token{
    fileName := fmt.Sprintf("%s.json", p.Auth.ClientId)

    tokenFile, err := os.Open(fileName) 
    defer tokenFile.Close()
    if err != nil{
        p.State = randomString(10) 
        callbackUrl := fmt.Sprintf("https://oauth.pipedrive.com/oauth/authorize?client_id=%s&state=%s&redirect_uri=%s", p.Auth.ClientId, p.State, p.RedirectUri)
        log.Println("No se pudo abrir el archivo", fileName)
        log.Println("Autoriza la aplicacion: ", callbackUrl)
        return nil
    }
    bytes, _ := ioutil.ReadAll(tokenFile)

    var token Token 
    err = json.Unmarshal(bytes, &token)
    if err != nil{
        log.Fatal("No se pudo leer el archivo", fileName)
    }
    p.Token = &token
    return &token
}

func (p *Pipedrive) ExchangeCodeToToken(code string) *Token{
    payload := map[string]string{
		"grant_type": "authorization_code",
		"code": code,
		"redirect_uri": p.RedirectUri,
    }
    
    p.Token = p.tokenApiCall(payload)
    if p.Token != nil{
        p.saveToken()
    }
    return p.Token
}

func (p *Pipedrive) refreshToken() *Token{
    if time.Now().Unix() > p.Token.ExpiresIn{
        if p.Token == nil{
            log.Fatal("No se puede refrescar un token que no existe")
        }

        log.Println("Refrescando token")
        payload := map[string]string{
            "grant_type": "refresh_token",
            "refresh_token": p.Token.RefreshToken,
        }

        p.Token = p.tokenApiCall(payload)
        if p.Token != nil{
            p.saveToken()
        }
    }
    return p.Token
}

func (p *Pipedrive) createDeal(c *models.Communication, userId uint32, personId uint32) (map[string]interface{}, error){
    title := c.Propiedad.Titulo
    if title == "" {
        title = "Trato con " + c.Nombre
    }

    payload := map[string]interface{}{
        "title": title,
        "user_id": userId,
        "person_id": personId,
        "value": c.Propiedad.Precio,
        "currency": "MXN",
        "visible_to": "3", //Para todos
    }

    var deal map[string]interface{}
    err := p.makeRequest("POST", "deals", payload, &deal)

    if err != nil{
        return nil, err
    }
    return deal, nil
}

func (p *Pipedrive) createPerson(c *models.Communication, ownerId uint32) (*Person, error){
    payload := map[string]interface{}{
        "name": c.Nombre,
        "owner_id": ownerId,
        "visible_to": "3", //visible para todos
        "phone": map[string]interface{}{
            "value": c.Telefono,
            "primary": true,
        },
        customFields["fuente"]: fuenteOptions[c.Fuente],
        customFields["fecha lead"]: c.FechaLead,
        customFields["link"]: c.Propiedad.Link,
        customFields["ubicacion"]: c.Propiedad.Ubicacion,
        customFields["precio"]: c.Propiedad.Precio,
    }
    fmt.Printf("payload: %v", payload)

    if (c.Email != ""){
        payload["email"] = map[string]interface{}{
            "value": c.Email,
            "primary": true,
        }
    }

    var person Person
    err := p.makeRequest("POST", "persons", payload, &person)

    if err != nil{
        return nil, err
    }
    return &person, nil
}

func (p *Pipedrive) getPersonByNumber(number string) (*Person, error){
    url := fmt.Sprintf("persons/search?term=%s&fields=%s", number, "phone")

    var data struct{
        Items []struct{
            Item Person `json:"item"`
        } `json:"items"`
    }
    err := p.makeRequest("GET", url, nil, &data)

    if err != nil{
        return nil, err
    }
    if len(data.Items) == 0{
        return nil, fmt.Errorf("La persona con telefono %s no existe", number)
    }

    return &data.Items[0].Item, nil
}

func (p *Pipedrive) getUserByName(name string) (*User, error){
    var users []User
    url := fmt.Sprintf("users/find?term=%s&search_by_email=0", name)

    err := p.makeRequest("GET", url, nil, &users)

    if err != nil || len(users) == 0{
        return nil, fmt.Errorf("El usuario con nombre: %s, no existe\n", name)
    }

    return &users[0], nil
}

func (p *Pipedrive) getField(id string) (*FieldOption, error){
    var field *FieldOption

    url := fmt.Sprintf("personFields/%s", id)
    err := p.makeRequest("GET", url, nil, field)

    if err != nil{
        return nil, fmt.Errorf("El field: %s, no existe\n", id)
    }

    return field, nil
}

func (p *Pipedrive) makeRequest(method string, path string, payload interface{}, data any) (error){
    p.refreshToken()
    url := baseUrl + path
    log.Println(url)
    
    var dataPayload io.Reader = nil
    if payload != nil {
        postBody, err := json.MarshalIndent(payload, "", "\t")
        if err != nil{
            return fmt.Errorf("No se pudo parsear el payload %s", err)
        }
        dataPayload = bytes.NewBuffer(postBody)
    }

    req, err := http.NewRequest(method, url, dataPayload)
    if err != nil {
        return err
    }

    req.Header.Add("Accept", "application/json")
    req.Header.Add("Content-Type", "application/json") 
    req.Header.Add("x-api-token", p.ApiToken)

    res, err := p.Client.Do(req)
    if err != nil {
        return err
    }
    defer res.Body.Close()

    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return err
    }

    jsonRes := Response{
        Success: true,
        Data: data,
    }
    err = json.Unmarshal(body, &jsonRes)
    if err != nil{
        return err
    }
    data = &jsonRes.Data

    if !jsonRes.Success{
        var a interface{}
        _ = json.Unmarshal(body, &a)
        return fmt.Errorf("La peticion no tuvo exito \nRes: %v\n", a)
    }
    return nil
}

func (p *Pipedrive) tokenApiCall(payload map[string]string) *Token{
    data := url.Values{}
    url := "https://oauth.pipedrive.com/oauth/token"

    for k, v := range payload{
        data.Set(k, v)
    }

    req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(data.Encode()))
    if err != nil{
        log.Printf("No se pudo construir la request", err)
        return nil
    }

    req.Header.Add("Authorization", p.Auth.Value) 
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded") 
    req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

    res, err := p.Client.Do(req) 
    if err != nil{
        log.Printf("Error en la request", err)
        return nil
    }
    defer res.Body.Close()

    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return nil
    }

    var token Token
    err = json.Unmarshal(body, &token)
    if err != nil {
        var debug interface{}
        _ = json.Unmarshal(body, &debug)
        log.Printf("La respuesta no matchea el token esperado %v %s", debug, err)
        return nil
    }
    fmt.Printf("Token: %v\n", token)
    log.Println("Token obtenido con exito")
    return &token
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
func randomString(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}
