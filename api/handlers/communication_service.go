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
    RoundRobin  *roundrobin.RoundRobin[models.Asesor]
    Logger  *slog.Logger

    Leads   *LeadService

    Flows   flow.FlowManager
    Utms    store.UTMStorer
    Comms   store.CommunicationStorer
    Store   store.Store
}

func (s CommunicationService) StoreCommunication(c *models.Communication) error {
	source, err := s.Store.GetSource(c)
	if err != nil {
		return err
	}

	lead, err := s.Leads.GetOrInsert(s.RoundRobin, c)
	if err != nil {
		return err
	}

    s.findUtmInMessage(c)
    
	c.Asesor = lead.Asesor
    c.Cotizacion = lead.Cotizacion
    c.Email = lead.Email
    c.Nombre = lead.Name

    // Insert the communication
    if err = s.Comms.Insert(c, source); err != nil {
        s.Logger.Error(err.Error(), "path", "InsertCommunication")
        return err
    }

    // Insert the communication
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

// GetOrInsert get the lead with phone c.Telefono in case that exists, otherwise creates one
// Must be in the CommunicationService
func (s *CommunicationService) saveLead(c *models.Communication) (*models.Lead, error) {
	lead, err := s.Leads.storer.GetOne(c.Telefono)

    // The lead does not exists
    if _, isStoreErr := err.(store.StoreError); isStoreErr{
		c.IsNew = true
        c.Asesor = *s.RoundRobin.Next()

		return s.Leads.storer.Insert(&models.CreateLead{
			Name:        c.Nombre,
			Phone:       c.Telefono,
			Email:       c.Email,
			AsesorPhone: c.Asesor.Phone,
            Cotizacion:  c.Cotizacion,
		})
	} else if lead != nil { // Duplicated lead
        c.Asesor = lead.Asesor

        updateLead := models.UpdateLead{
            Name: c.Nombre,
            Cotizacion: c.Cotizacion,
            Email: c.Email,
        }

        updateLeadsFields(lead, updateLead)
		return lead, s.Leads.storer.Update(lead)
    }

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
    }else{
        go s.Flows.RunMainFlow(c)
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
                Medium: utm.Medium,
                Source: utm.Source,
                Campaign: utm.Campaign,
                Ad: utm.Ad,
                Channel: utm.Channel,
            }
        }
    }
}
