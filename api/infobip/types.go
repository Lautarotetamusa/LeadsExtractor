package infobip

import (
	"log/slog"
	"net/http"
)

type InfobipApi struct{
   apiUrl string
   apiKey string
   sender string
   client *http.Client
   logger *slog.Logger
}

type SendMessage struct{
    From    string      `json:"from"`
    To      string      `json:"to"`
    Content interface{} `json:"content"`
}

type TextContent struct{
    Text string `json:"text"`
}
type MediaContent struct{
    MediaUrl string `json:"mediaUrl"`
}
type TemplateContent struct{
    TemplateName string `json:"template_name"`
    Language     string `json:"language"`   
}
type SendTemplatesPayload struct{
    Messages []SendMessage `json:"messages"`
}

type PhoneContact struct{
    Number string `json:"number"`
}
type EmailContact struct{
    Address string `json:"address,omitempty"`
}

type ContactInformation struct {
    Phone *PhoneContact `json:"phone,omitempty"`
    Email *EmailContact `json:"email,omitempty"`
}

type CustomAttributes struct {
    PropLink		string  `json:"prop_link"`
    PropPrecio		string  `json:"prop_precio"`
    PropUbicacion	string  `json:"prop_ubicacion"`
    PropTitulo		string  `json:"prop_titulo"`
    Contacted		bool    `json:"contacted"`
    Fuente		    string  `json:"fuente"`
    AsesorName		string  `json:"asesor_name"`
    AsesorPhone	    string  `json:"asesor_phone"`
    FechaLead		string  `json:"fecha_lead"`
}

type InfobipLead struct {
    FirstName   string `json:"firstName"`
    LastName    string `json:"lastName"`
    Type        string `json:"type"`
    CustomAttributes    CustomAttributes    `json:"customAttributes"`
    ContactInformation  ContactInformation  `json:"contactInformation"`
    Tags                string              `json:"tags"`
}
