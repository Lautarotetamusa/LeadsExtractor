package jotform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"leadsextractor/models"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Form struct {
    id  string
}

type Jotform struct {
    apiKey  string    
    forms   []Form
    apiHost string
}

func NewJotform(apiKey string, apiHost string) *Jotform {
    return &Jotform{
        apiKey: apiKey,
        apiHost: apiHost,
        forms: make([]Form, 0),
    }
}

func (j *Jotform) AddForm(formId string) *Form {
    form := Form{
        id: formId,
    }
    j.forms = append(j.forms, form)
    return &form
}

//Hace una peticion a la app que genera un pdf de jotform
func (j *Jotform) GetPdf(c *models.Communication, f *Form) (string, error) {
    payload := map[string]any{
        "data": c,
    }
	jsonBody, _ := json.Marshal(payload)
	bodyReader := bytes.NewReader(jsonBody)

    url := fmt.Sprintf("%s/jotform", j.apiHost)

	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		return "", err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

    client := &http.Client{
        Timeout: 30 * time.Second,
    }

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("no se pudo realizar la peticion: %s", err)
	}
    
	defer res.Body.Close()
    buf, err := io.ReadAll(res.Body)
    if err != nil {
        return "", err
    }

    if res.StatusCode != http.StatusOK {
        return "", fmt.Errorf("request not ok, status = %d", res.StatusCode)
    }

    return string(buf), nil
}

func (j *Jotform) SubmitForm(c *models.Communication, f *Form) error {
    if c.Asesor.Email == ""{
        return fmt.Errorf("el asesor %s no tiene email asignado", c.Asesor.Name)
    }

    if c.Busquedas.CoveredArea.String == "" || !c.Busquedas.CoveredArea.Valid {
        c.Busquedas.CoveredArea.String = "350, 350"
    }

    if c.Busquedas.Banios.String == "" || !c.Busquedas.Banios.Valid {
        c.Busquedas.Banios.String = "5, 5"
    }

    if c.Busquedas.Recamaras.String == "" || !c.Busquedas.Recamaras.Valid {
        c.Busquedas.Recamaras.String = "4, 4"
    }

    email := "cotizaciones@rebora.com.mx"
    if !c.Email.Valid {
        email = c.Email.String
    }

    props := map[string]float32{
        "covered_area": parseProp(c.Busquedas.CoveredArea.String),
        "banios":       parseProp(c.Busquedas.Banios.String),
        "recamaras":    parseProp(c.Busquedas.Recamaras.String),
    }

    params := map[string]string{
        "nombreCliente": c.Nombre,
        "emailCliente": email,
        "escribaUna": c.Asesor.Name,
        "email123": c.Asesor.Email,
        "numeroDe[full]": c.Asesor.Phone,
        "pagoInicial": "2,500,000",
        "pagoInicial118": "25%",
        "aCuantos": "16+meses",
        "escribaUna9": "Premium",
        "cuantosCuartos": strconv.Itoa(int(props["recamaras"])),
        "cuantosBanos56": strconv.Itoa(int(props["banios"])),
        "cuantosBanos": "0", 
        "tamanoDe":     strconv.Itoa(int(props["covered_area"] * 0.45)),   //Tamaño planta baja
        "tamanoDe79":   strconv.Itoa(int(props["covered_area"] * 0.45)), //Tamaño planta alta
        "tamanoDel":    strconv.Itoa(int(props["covered_area"] * 0.10)),  //Tamaño roof garden
        "tamanoDe84": "0",  //Tamaño rampa estacionamiento
        "tamanoDe85": "30", //Tamaño jardin exterior
        "tamanoDe86": "0",  //Tamaño de alberca
        "tamanoDe87": "50", //Tamaño muro perimetral
    }

    fmt.Printf("%v\n", params)

    url := fmt.Sprintf("https://api.jotform.com/form/%s/submissions", f.id)
    res, err := j.sendRequest(url, params); 
    if err != nil {
        return err
    }

    fmt.Printf("%v\n", res)

    return nil
}

func (j *Jotform) sendRequest(url string, payload map[string]string) (any, error) {
	jsonBody, _ := json.Marshal(payload)
	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
    req.Header.Add("APIKEY", j.apiKey)

    client := &http.Client{
        Timeout: 40 * time.Second,
    }

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("no se pudo realizar la peticion: %s", err)
	}
    
	defer res.Body.Close()
	var response any
    if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
        return nil, err
    }

	return response, nil
}

func parseProp(prop string) float32 {
    split := strings.Split(prop, ", ")
    min, _ := strconv.Atoi(split[0])
    max, _ := strconv.Atoi(split[1])
    return float32((min + max) / 2)
}
