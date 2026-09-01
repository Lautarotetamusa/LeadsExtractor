package handlers

import (
	"fmt"
	"leadsextractor/flow"
	"leadsextractor/models"
	"leadsextractor/pkg/roundrobin"
	"leadsextractor/store"
	"log/slog"
	"strings"
)

type CommunicationService struct {
	RoundRobin *roundrobin.RoundRobin[models.Asesor]
	Logger     *slog.Logger

	Leads      store.LeadStorer
	Flows      flow.FlowManager
	Utms       store.UTMStorer
	Comms      store.CommunicationStorer
	Properties store.PropertyStorer
	Store      store.Storer
}

func (s CommunicationService) StoreCommunication(c *models.Communication) error {
	if err := s.attachProperty(c); err != nil {
		return err
	}

	_, err := s.SaveLead(c)
	if err != nil {
		return err
	}

	s.findUtmInMessage(c)

	// Insert the communication
	if err = s.Comms.Insert(c); err != nil {
		s.Logger.Error(err.Error(), "path", "InsertCommunication")
		return err
	}

	// Insert the message
	if c.Message.Valid && c.Message.String != "" {
		if err = s.Store.InsertMessage(store.CommunicationToMessage(c)); err != nil {
			return err
		}
	}
	return nil
}

func (s CommunicationService) NewCommunication(c *models.Communication) error {
	err := s.StoreCommunication(c)
	if err != nil {
		return err
	}

	if text, err := s.Store.GetLastSentMessageToLead(c.Telefono); err == nil {
		c.LastSentMessage = models.NullString{String: text, Valid: true}
	}

	s.runAction(c)

	return nil
}

// attachProperty resuelve la Property asociada a la comunicación cuando
// viene de un portal (busca por portal_id+portal y la crea si no existe) y
// deja c.Propiedad.ID seteado para que Comms.Insert la guarde en
// Communication.property_id. Para whatsapp/ivr/viewphone no hace nada:
// esas comunicaciones no tienen Property asociada.
func (s CommunicationService) attachProperty(c *models.Communication) error {
	if err := store.ValidateSource(c.Fuente); err != nil {
		return err
	}

	if !store.FuenteTienePropiedad(c.Fuente) {
		return nil
	}

	property, _ := s.Properties.GetProperty(c.Propiedad.PortalId.String, c.Fuente)
	if property != nil {
		c.Propiedad.ID = property.ID
		return nil
	}

	c.Propiedad.Portal = c.Fuente
	return s.Properties.InsertProperty(&c.Propiedad)
}

// saveLead get the lead with phone c.Telefono in case that exists, otherwise creates one
func (s *CommunicationService) SaveLead(c *models.Communication) (*models.Lead, error) {
	lead, err := s.Leads.GetOne(c.Telefono)

	if err != nil {
		// Another error, not "lead not found"
		if _, isStoreErr := err.(store.StoreError); !isStoreErr {
			return nil, err
		}

		// The lead does not exists
		c.IsNew = true
		c.Asesor = *s.RoundRobin.Next()

		lead, err = s.Leads.Insert(&models.CreateLead{
			Name:        c.Nombre,
			Phone:       c.Telefono,
			Email:       c.Email,
			AsesorPhone: c.Asesor.Phone,
		})
		if err != nil {
			fmt.Printf("Error creating Lead: %#v\n", err)
		}
	}

	if lead != nil { // Duplicated lead
		c.Asesor = lead.Asesor

		updateLead := models.UpdateLead{
			Name:  c.Nombre,
			Email: c.Email,
		}

		updateLeadsFields(lead, updateLead)
		err = s.Leads.Update(lead)
	}

	// Update the communication Lead fields
	c.Email = lead.Email
	c.Nombre = lead.Name

	// another error in GetOne
	return lead, err
}

func (s CommunicationService) runAction(c *models.Communication) {
	action, err := s.Store.GetLastActionFromLead(c.Telefono)
	if err != nil {
		s.Logger.Warn("cannot get the lead last action", "err", err)
	}

	if action != nil && action.OnResponse.Valid {
		go s.Flows.RunFlow(c, action.OnResponse.UUID)
	} else {
		go s.Flows.RunMain(c)
	}
}

// find if the message have match with any utm
func (s CommunicationService) findUtmInMessage(c *models.Communication) {
	utms, err := s.Utms.GetAll()
	if err != nil {
		s.Logger.Error(err.Error())
		return
	}
	if !c.Message.Valid {
		return
	}

	message := strings.ToUpper(c.Message.String)
	for _, utm := range utms {
		if strings.Contains(message, utm.Code) {
			s.Logger.Info(fmt.Sprintf("found code %s in message", utm.Code))
			c.Utm = models.Utm{
				Medium:   utm.Medium,
				Source:   utm.Source,
				Campaign: utm.Campaign,
				Ad:       utm.Ad,
				Channel:  utm.Channel,
			}
		}
	}
}
