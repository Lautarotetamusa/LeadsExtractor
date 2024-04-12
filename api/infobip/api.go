package infobip

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"leadsextractor/models"
	"log"
	"net/http"
	"net/url"
	"os"
)

func NewInfobipApi() *InfobipApi{
    return &InfobipApi{
        apiUrl: os.Getenv("INFOBIP_APIURL"),
        apiKey: os.Getenv("INFOBIP_APIKEY"),
        sender: "5213328092850",
        client: &http.Client{},
    }
}

func (i *InfobipApi) SendWppTemplate(to string, templateName string, language string){
    if language == ""{
        language = "es_MX"
    }

    fmt.Printf("Enviando %s template por whatsapp", templateName)
    message := SendMessage{
        From: i.sender,
        To: to,
        Content: TemplateContent{
            TemplateName: templateName,
            Language: language,
        },
    }
    payload := &SendTemplatesPayload{
        Messages: []SendMessage{message},
    }
    
    go i.makeRequest("POST", "/whatsapp/1/message/template", payload)
}

func (i *InfobipApi) SendWppMediaMessage(to string, tipo string, mediaUrl string){
    fmt.Printf("Enviando %s por whatsapp\n", tipo)
    payload := &SendMessage{
        From: i.sender,
        To: to,
        Content: MediaContent{
            MediaUrl: mediaUrl,
        },
    }

    url := fmt.Sprintf("/whatsapp/1/message/%s", tipo)
    go i.makeRequest("POST", url, payload)
}

func (i *InfobipApi) SendWppTextMessage(to string, message string){
    log.Println("Enviando mensaje de texto por whatsapp")
    payload := &SendMessage{
        From: i.sender,
        To: to,
        Content: TextContent{
            Text: message,
        },
    }

    go i.makeRequest("POST", "/whatsapp/1/message/text", payload)
}

func (i *InfobipApi) SaveLead(lead *InfobipLead){
    log.Println("Cargando lead a infobip")
    go i.makeRequest("POST", "/people/2/persons", lead)
}

func Communication2Infobip(c *models.Communication) *InfobipLead{
    var contactInformation ContactInformation 
    if c.Email != ""{
        contactInformation.Email = &EmailContact{
            Address: c.Email,
        }
    }
    if c.Telefono != ""{
        contactInformation.Phone = &PhoneContact{
            Number: c.Telefono,
        }
    }

    return &InfobipLead{
        FirstName: c.Nombre,
        LastName: "",
        Type: "LEAD",
        CustomAttributes: CustomAttributes{
            PropLink: c.Propiedad.Link,
            PropPrecio: c.Propiedad.Precio,
            PropUbicacion: c.Propiedad.Ubicacion,
            PropTitulo: c.Propiedad.Titulo,
            Contacted: false,
            Fuente: c.Fuente,
            AsesorName: c.Asesor.Name,
            AsesorPhone: c.Asesor.Phone,
            FechaLead: c.FechaLead,
        },
        ContactInformation: contactInformation,
        Tags: "Seguimientos",
    }
}

func (i *InfobipApi) makeRequest(method string, path string, payload interface{}){
    postBody, err := json.MarshalIndent(payload, "", "\t")
    if err != nil{
        log.Printf("No se pudo parsear el payload", err)
        return
    }
    fmt.Println("body:", string(postBody))
    data := bytes.NewBuffer(postBody)

    url, err := url.JoinPath(i.apiUrl, path)
    if err != nil{
        log.Printf("La url es incorrecta", path)
        return
    }

    req, err := http.NewRequest(method, url, data)
    if err != nil{
        log.Printf("No se pudo construir la request", err)
        return
    }

    req.Header.Add("Authorization", i.apiKey) 
    req.Header.Add("Content-Type", "application/json") 

    res, err := i.client.Do(req) 
    if err != nil{
        log.Printf("Error en la request", err)
        return
    }

    defer res.Body.Close()
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        log.Printf("Error obteniendo la respuesta", err)
        return
    }

    log.Printf("Request enviada correctamente", string(body))
}
