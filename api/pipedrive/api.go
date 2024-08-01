package pipedrive

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"leadsextractor/models"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const baseUrl = "https://api.pipedrive.com/v1/"

type Pipedrive struct{
    token       *Token 
    auth        Auth
    apiToken    string
    redirectUri string
    client      *http.Client
    logger      *slog.Logger
    filePath    string
    State       string
}

type Config struct {
    ClientId        string
    ClientSecret    string
    ApiToken        string
    RedirectURI     string
    FilePath        string
}

type Token struct{
    AccessToken     string `json:"access_token,omitempty"`
    RefreshToken    string `json:"refresh_token,omitempty"`
    Scope           string `json:"scope,omitempty"`
    ExpiresIn       int64  `json:"expires_in,omitempty"` //Tiempo Unix() en el que expira el token
    ApiDomain       string `json:"api_domain,omitempty"`
}

type Auth struct{
    clientId        string
    clientSecret    string
    value           string
}

type Response struct{
    Success bool    `json:"success"`
    Data    any     `json:"data"`
}

type Pagination struct {
    Start   int                 `json:"start"`
    Limit	int                 `json:"limit"`
    MoreItemsInCollection bool  `json:"more_items_in_collection"`
    NextStart int               `json:"next_start"` 
}

type AdditionalData struct {
    Pagination *Pagination `json:"pagination"`
}

type PaginatedResponse struct {
    Success bool                    `json:"success"`
    Data            []any           `json:"data"`
    AdditionalData  *AdditionalData `json:"additional_data"`
}

func NewPipedrive(c Config, l *slog.Logger) *Pipedrive{
    client := fmt.Sprintf("%s:%s", c.ClientId, c.ClientSecret) 
    p := Pipedrive{
        auth: Auth{
            clientId: c.ClientId,
            clientSecret: c.ClientSecret,
            value: fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(client))),
        },
        client: &http.Client{
            Timeout: 15 * time.Second,
        },
        redirectUri: c.RedirectURI,
        apiToken: c.ApiToken,
        token: nil,
        logger: l.With("module", "pipedrive"),
        filePath: c.FilePath,
    }
    if p.filePath == "" {
        p.filePath = "./"
    }
    p.loadToken()

    if p.token != nil{
        p.refreshToken()        
    }
    return &p
}

func (p *Pipedrive) HandleOAuth(w http.ResponseWriter, r *http.Request) error {
	code := r.URL.Query().Get("code")
	log.Println("Code:", code)
	p.ExchangeCodeToToken(code)
	w.Write([]byte(p.State))
	return nil
}

func (p *Pipedrive) SaveCommunication(c *models.Communication){
    asesor, err := p.GetUserByEmail(c.Asesor.Email)
    if err != nil{
        p.logger.Error("Error obteniendo el asesor %s", "err", err)
        return
    }
    p.logger.Debug(fmt.Sprintf("Asesor: %v", asesor))

    person, err := p.GetPersonByNumber(c.Telefono)
    if err != nil{
        p.logger.Warn("No se encontro al asesor", "err", err)
        p.logger.Debug("Creando asesor en PipeDrive")
        person, err = p.createPerson(c, asesor.Id)
        if err != nil{
            p.logger.Error("Error creando persona", "err", err)
            return
        }
    }
    p.logger.Debug(fmt.Sprintf("Person: %v", person))

    deal, err := p.SearchPersonDeal(person.Id, asesor.Id)
    if err != nil{
        p.logger.Warn("No se encontro al deal", "err", err)
        p.logger.Debug("Cargando deal en PipeDrive")
        deal, err = p.createDeal(c, asesor.Id, person.Id)
        if err != nil{
            p.logger.Error("Error creando deal", "err", err)
            return
        }

        p.logger.Info("Deal cargando correctamente en PipeDrive")
    }else{
        p.logger.Info("El Deal ya estaba cargado")
    }
    p.logger.Debug(fmt.Sprintf("Deal: %v", deal))

    _, err = p.addNote(c, deal.Id)
    if err != nil {
        p.logger.Error("Error cargando la nota", "err", err)
    }else{
        p.logger.Info("Nota cargada con exito")
    }
}

func (p *Pipedrive) saveToken(){
    //Hacemos que el campo expiresIn sea el tiempo Unix en el que expirará
    p.token.ExpiresIn = time.Now().Unix() + int64(p.token.ExpiresIn)

    file, _ := json.MarshalIndent(p.token, "", "\t")
    fileName := fmt.Sprintf("%s%s.json", p.filePath, p.auth.clientId)

    err := os.WriteFile(fileName, file, 0644)
    if err != nil{
        log.Fatal("No se pudo guardar el token en el archivo")
    }
}

func (p *Pipedrive) loadToken() {
    fileName := fmt.Sprintf("%s%s.json", p.filePath, p.auth.clientId)

    tokenFile, err := os.Open(fileName) 
    if err != nil{
        p.State = randomString(10) 
        callbackUrl := fmt.Sprintf("https://oauth.pipedrive.com/oauth/authorize?client_id=%s&state=%s&redirect_uri=%s", p.auth.clientId, p.State, p.redirectUri)
        p.logger.Warn("No se pudo abrir el archivo", "file", fileName)
        log.Println("Autoriza la aplicacion: ", callbackUrl)
        //TODO: Mantener la coneccion durante un tiempo X hasta que la aplicacion sea autorizada
        //Se puede hacer con un channel en la funcion de callback
    }
    defer tokenFile.Close()
    bytes, _ := io.ReadAll(tokenFile)

    var token Token 
    err = json.Unmarshal(bytes, &token)
    if err != nil{
        log.Fatal("No se pudo leer el archivo", fileName)
    }
    p.token = &token
}

func (p *Pipedrive) ExchangeCodeToToken(code string) *Token{
    payload := map[string]string{
		"grant_type": "authorization_code",
		"code": code,
		"redirect_uri": p.redirectUri,
    }
    
    p.token = p.tokenApiCall(payload)
    if p.token != nil{
        p.saveToken()
    }
    return p.token
}

func (p *Pipedrive) refreshToken() *Token{
    if time.Now().Unix() > p.token.ExpiresIn{
        if p.token == nil{
            log.Fatal("No se puede refrescar un token que no existe")
        }

        p.logger.Info("Refrescando token")
        payload := map[string]string{
            "grant_type": "refresh_token",
            "refresh_token": p.token.RefreshToken,
        }

        p.token = p.tokenApiCall(payload)
        if p.token != nil{
            p.saveToken()
        }
    }
    return p.token
}

func (p *Pipedrive) makeRequest(method string, path string, payload any, data any) error {
    p.refreshToken()
    url := baseUrl + path
    
    var dataPayload io.Reader = nil
    if payload != nil {
        postBody, err := json.MarshalIndent(payload, "", "\t")
        if err != nil{
            return fmt.Errorf("no se pudo parsear el payload %s", err)
        }
        dataPayload = bytes.NewBuffer(postBody)
    }

    req, err := http.NewRequest(method, url, dataPayload)
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
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
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
        return fmt.Errorf("la peticion no tuvo exito \nRes: %v", a)
    }
    return nil
}

func (p *Pipedrive) getPaginatedData(url string, data *[]any) error {
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

        var response PaginatedResponse
        if err = json.NewDecoder(res.Body).Decode(&response); err != nil{
            return err
        }

        moreItems = response.AdditionalData.Pagination.MoreItemsInCollection
        start = response.AdditionalData.Pagination.NextStart

        *data = append(response.Data, data)
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
        p.logger.Error("No se pudo construir la request", "err", err)
        return nil
    }

    req.Header.Add("Authorization", p.auth.value) 
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded") 
    req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

    res, err := p.client.Do(req) 
    if err != nil{
        p.logger.Error("Error en la request", "err", err)
        return nil
    }
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    if err != nil {
        return nil
    }

    var token Token
    err = json.Unmarshal(body, &token)
    if err != nil {
        var debug interface{}
        _ = json.Unmarshal(body, &debug)
        p.logger.Error(fmt.Sprintf("La respuesta no matchea el token esperado %v %s", debug, err))
        return nil
    }
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
