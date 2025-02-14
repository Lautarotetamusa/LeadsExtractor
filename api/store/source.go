package store

import (
	"database/sql"
	"fmt"
	"leadsextractor/models"
	"slices"
	"strings"

	"github.com/jmoiron/sqlx"
)

type SourceStorer interface {
    InsertSource(propertyId int) error
    GetSource(sourceType string) (*models.Source, error)
    GetPropertySource(propertyId int32) (*models.Source, error)
    // Propiedad.ID must be setted after the insert
    InsertProperty(*models.Propiedad) error
    GetProperty(portalId string, portal string) (*models.Propiedad, error)
}

type SourceDBStore struct {
    db *sqlx.DB
}

func NewSourceDBStore(db *sqlx.DB) *SourceDBStore {
    return &SourceDBStore{
        db: db,
    }
}

const (
    insertPropQ = `INSERT INTO Property (portal_id, title, url, price, ubication, tipo, portal) 
                   VALUES (:portal_id, :title, :url, :price, :ubication, :tipo, :portal)`
    
    selectPropQ = "SELECT * FROM Property WHERE portal_id LIKE ? AND portal = ? LIMIT 1"

    getPropertySourceQ = "SELECT * FROM Source WHERE property_id=?"

    getSourceQ = "SELECT * FROM Source WHERE type LIKE ?"

    insertSourceQ =  "INSERT INTO Source (type, property_id) VALUES(:type, :property_id)"
)

var validSources = []string{"whatsapp", "ivr", "viewphone", "inmuebles24", "lamudi", "casasyterrenos", "propiedades"}

func ValidateSource(source string) error {
    if !slices.Contains(validSources, source){
		return fmt.Errorf("source its not valid, must be one of (%s)", source, strings.Join(validSources, ", "))
    }
    return nil
}

func (s *SourceDBStore) GetPropertySource(propertyId int32) (*models.Source, error) {
	source := models.Source{}
    err := s.db.Get(&source, getPropertySourceQ, propertyId)
	if err != nil {
        return nil, fmt.Errorf("error obteniendo el source con property_id: %d", propertyId)
	}
    return &source, nil
}

func (s *SourceDBStore) GetSource(sourceType string) (*models.Source, error) {
	source := models.Source{}
    err := s.db.Get(&source, getSourceQ, sourceType)
    if err != nil {
        return nil, fmt.Errorf("source: %s does not exists", sourceType)
    }

    return &source, nil
}

func (s *SourceDBStore) GetProperty(portalId string, portal string) (*models.Propiedad, error) {
	property := models.Propiedad{}
	err := s.db.Get(&property, selectPropQ, portalId, portal)
    if err != nil {
        return nil, fmt.Errorf("the property with id %s doest not exists", portalId)
    }
    return &property, nil
}

func (s *SourceDBStore) InsertSource(propertyId int) error {
    source := models.Source{
        Tipo:       "property",
        PropertyId: sql.NullInt16{Int16: int16(propertyId), Valid: true},
    }

    if _, err := s.db.NamedExec(insertSourceQ, source); err != nil {
        return fmt.Errorf("error inserting source: %s", err.Error())
    }
    return nil
}

func (s *SourceDBStore) InsertProperty(p *models.Propiedad) error {
    res, err := s.db.NamedExec(insertQuery, &p); 
    if err != nil {
        return fmt.Errorf("error inserting property: %s", err.Error())
    }

    pId, err := res.LastInsertId()
    p.ID = models.NullInt32{Int32: int32(pId), Valid: true}
    return nil
}
