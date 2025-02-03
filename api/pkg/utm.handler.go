package pkg

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"leadsextractor/models"
	"net/http"
	"strconv"
	"unicode"

	"github.com/gorilla/mux"
)

var validChannels = []string{"ivr", "whatsapp", "inbox"}
func validateChannel(utm *models.UtmDefinition) error {
    if !slices.Contains(validChannels, utm.Channel.String) {
        return fmt.Errorf("el channel no es valido. validos: %s", strings.Join(validChannels, ", "))
    }
    return nil
}

func (s *Server) GetAllUtmes(w http.ResponseWriter, r *http.Request) error {
    var utms []models.UtmDefinition
	if err := s.Store.GetAllUtm(&utms); err != nil {
		return err
	}

	dataResponse(w, utms)
	return nil
}

func (s *Server) GetOneUtm(w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["id"]
    id, err := strconv.Atoi(idStr)
    utm, err := s.Store.GetOneUtm(id)
	if err != nil {
        return fmt.Errorf("no se encontro el utm con id %d", id)
	}

	dataResponse(w, utm)
	return nil
}

func (s *Server) InsertUtm(w http.ResponseWriter, r *http.Request) error {
	var utm models.UtmDefinition
    defer r.Body.Close()
    if err := json.NewDecoder(r.Body).Decode(&utm); err != nil {
        return err
    }

    if err := validateChannel(&utm); err != nil {
        return err
    }

    isValid := true
    for _, r := range utm.Code {
        if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
            isValid = false
        }
    }

    if !isValid {
        return fmt.Errorf("el codigo solo puede contener caracteres alphanumericos")
    }
    utm.Code = strings.ToUpper(utm.Code)

    id, err := s.Store.InsertUtm(&utm); 
    if err != nil {
		return err
	}
    utm.Id = int(id)

	successResponse(w, "Utm creado correctamente", utm)
	return nil
}

func (s *Server) UpdateUtm(w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["id"]
    id, err := strconv.Atoi(idStr)
    utm, err := s.Store.GetOneUtm(id)
    if err != nil {
        return fmt.Errorf("no existe utm con id %d", id)
    }

    defer r.Body.Close()
    if err := json.NewDecoder(r.Body).Decode(&utm); err != nil {
		return err
    }

    if err := validateChannel(utm); err != nil {
        return err
    }

    if err := s.Store.UpdateUtm(utm); err != nil {
		return err
	}

	successResponse(w, "Utm actualizado correctamente", utm)
	return nil
}
