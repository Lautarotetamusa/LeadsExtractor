package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const baseUrl = "https://graph.facebook.com/v17.0/%s/messages"

type Whatsapp struct{
    accessToken string 
    numberId    string
    imageId     string
    videoId     string
    client      *http.Client
    url         string
}

type Response struct{
    MessagingProduct string `json:"messaging_product"`
    Contacts []struct{
        Input   string `json:"input"`
        WaId    string `json:"wa_id"`
    }
    Messages []struct{
        Id string `json:"id"`
    }
}

type TextPayload struct {
    PreviewUrl  bool    `json:"preview_url"`
    Body        string  `json:"body"`
}

func NewWhatsapp(accesToken string, numberId string) *Whatsapp{
    w := &Whatsapp{
        client: &http.Client{
            Timeout: 15 * time.Second,
        },
        accessToken: accesToken,
        numberId: numberId,
        url: fmt.Sprintf(baseUrl, numberId),
    } 
    fmt.Printf("url: %s\n", w.url)
    return w
}

func (w *Whatsapp) request(payload interface{}) (*Response, error){
    jsonBody, err := json.Marshal(payload)
    bodyReader := bytes.NewReader(jsonBody)

    req, err := http.NewRequest(http.MethodPost, w.url, bodyReader)
    if err != nil {
        return nil, err
    }

    req.Header.Add("Accept", "application/json")
    req.Header.Add("Content-Type", "application/json") 
    req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", w.accessToken))

    res, err := w.client.Do(req)
    defer res.Body.Close()
    if err != nil {
        return nil, fmt.Errorf("No se pudo realizar la peticion: %s\n", err)
    }

    bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error al leer el cuerpo de la respuesta: %w", err)
	}

    var data Response
    err = json.Unmarshal(bodyBytes, &data)
    if len(data.Messages) == 0 {
        var debug interface{}
        _ = json.Unmarshal(bodyBytes, &debug)
        fmt.Printf("debug: %v", debug)
        return nil, fmt.Errorf("No se pudo obtener el json de la peticion: %s\n", err)
    }

    return &data, nil
}

func (w *Whatsapp) SendMessage(to string, message string) (*Response, error){
    payload := map[string]interface{}{
        "messaging_product": "whatsapp",
        "recipient_type": "individual",
        "to": to,
        "type": "text",
        "text": TextPayload{
            PreviewUrl: false,
            Body: message,
        },
    }
    return w.request(payload)
}

/*
func (w *Whatsapp) SendTemplate(to string, name string, components any, language string = "es_MX"){
    payload := map[string]interface{}{
        "messaging_product": "whatsapp",
        "recipient_type": "individual",
        "to": to,
        "type": "template",
        "template": { 
            "name": name, 
            "language": {
                "code": language,
            },
            "components": components,
        },
    }
    return w.request(payload)
}*/
