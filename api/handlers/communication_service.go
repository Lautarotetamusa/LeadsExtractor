package handlers

import (
	"fmt"
	"leadsextractor/flow"
	"leadsextractor/models"
	"leadsextractor/pkg/roundrobin"
	"leadsextractor/store"
	"log/slog"
	"slices"
	"strings"
)

type CommunicationService struct {
	RoundRobin *roundrobin.RoundRobin[models.Asesor]
	Logger     *slog.Logger

	Leads  store.LeadStorer
	Flows  flow.FlowManager
	Utms   store.UTMStorer
	Comms  store.CommunicationStorer
	Source store.SourceStorer
	Store  store.Storer
}

func (s CommunicationService) StoreCommunication(c *models.Communication) error {
	source, err := s.getOrInsertSource(c)
	if err != nil {
		return err
	}

	_, err = s.SaveLead(c)
	if err != nil {
		return err
	}

	s.findUtmInMessage(c)

	// Insert the communication
	if err = s.Comms.Insert(c, source); err != nil {
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

	s.runAction(c)

	return nil
}

func (s CommunicationService) getOrInsertSource(c *models.Communication) (*models.Source, error) {
	if err := store.ValidateSource(c.Fuente); err != nil {
		return nil, err
	}

	if slices.Contains([]string{"whatsapp", "ivr", "viewphone"}, c.Fuente) {
		return s.Source.GetSource(c.Fuente)
	}

	property, err := s.getOrInsertProperty(c)
	if err != nil {
		return nil, err
	}

	return s.Source.GetPropertySource(property.ID.Int32)
}

// If alredy exists a property with that portal_id in the db, retrieves its
// otherwise, creates a new property and insert a source with the generated prop id
func (s CommunicationService) getOrInsertProperty(c *models.Communication) (*models.Propiedad, error) {
	property, err := s.Source.GetProperty(c.Propiedad.PortalId.String, c.Fuente)

	// If the property doesnt exists, create it
	if property == nil {
		c.Propiedad.Portal = c.Fuente
		property = &c.Propiedad

		// Insert the property and get the property db auto generated id
		err = s.Source.InsertProperty(property)
		if err != nil {
			return nil, err
		}

		// Insert the source with the generated id
		err = s.Source.InsertSource(int(property.ID.Int32))
		if err != nil {
			return nil, err
		}
	}

	return property, err
}

// saveLead get the lead with phone c.Telefono in case that exists, otherwise creates one
func (s *CommunicationService) SaveLead(c *models.Communication) (*models.Lead, error) {
	lead, err := s.Leads.GetOne(c.Telefono)

	// The lead does not exists
	if _, isStoreErr := err.(store.StoreError); isStoreErr {
		c.IsNew = true
		c.Asesor = *s.RoundRobin.Next()

		lead, err = s.Leads.Insert(&models.CreateLead{
			Name:        c.Nombre,
			Phone:       c.Telefono,
			Email:       c.Email,
			AsesorPhone: c.Asesor.Phone,
			Cotizacion:  c.Cotizacion,
		})
	} else if lead != nil { // Duplicated lead
		c.Asesor = lead.Asesor

		updateLead := models.UpdateLead{
			Name:       c.Nombre,
			Cotizacion: c.Cotizacion,
			Email:      c.Email,
		}

		updateLeadsFields(lead, updateLead)
		err = s.Leads.Update(lead)
	}

	// Update the communication Lead fields
	c.Cotizacion = lead.Cotizacion
	c.Email = lead.Email
	c.Nombre = lead.Name

	// another error in GetOne
	return lead, err
}

func (s CommunicationService) runAction(c *models.Communication) {
	action, err := s.Store.GetLastActionFromLead(c.Telefono)
	if err != nil {
		s.Logger.Error("cannot get the lead last action", "err", err)
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
