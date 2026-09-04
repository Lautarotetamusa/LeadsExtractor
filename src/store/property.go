package store

import (
	"fmt"
	"leadsextractor/models"
	"slices"
	"strings"

	"github.com/jmoiron/sqlx"
)

// PropertyStorer resuelve la Property asociada a una comunicación que vino
// de un portal: la busca por (portal_id, portal) y la crea si no existe.
// Las comunicaciones de whatsapp/ivr/viewphone no tienen Property asociada.
type PropertyStorer interface {
	InsertProperty(*models.Propiedad) error
	GetProperty(portalId string, portal string) (*models.Propiedad, error)
}

type PropertyDBStore struct {
	db *sqlx.DB
}

func NewPropertyDBStore(db *sqlx.DB) *PropertyDBStore {
	return &PropertyDBStore{
		db: db,
	}
}

const (
	insertPropQ = `INSERT INTO Property (portal_id, title, url, price, ubication, tipo, portal, bedrooms, bathrooms, total_area, covered_area)
                   VALUES (:portal_id, :title, :url, :price, :ubication, :tipo, :portal, :bedrooms, :bathrooms, :total_area, :covered_area)`

	selectPropQ = "SELECT * FROM Property WHERE portal_id LIKE ? AND portal = ? LIMIT 1"
)

var validSources = []string{"whatsapp", "ivr", "viewphone", "inmuebles24", "lamudi", "casasyterrenos", "propiedades", "easybroker", "wordpress"}

// fuentesSinPropiedad son los orígenes de una comunicación que no están
// ligados a ninguna Property.
var fuentesSinPropiedad = []string{"whatsapp", "ivr", "viewphone"}

func ValidateSource(source string) error {
	if !slices.Contains(validSources, source) {
		return fmt.Errorf("source %q its not valid, must be one of (%s)", source, strings.Join(validSources, ", "))
	}
	return nil
}

// ValidSources devuelve una copia de las fuentes válidas de una
// comunicación (ver ValidateSource) — usado por los tests de integración
// para ejercitar cada una contra una DB real.
func ValidSources() []string {
	out := make([]string, len(validSources))
	copy(out, validSources)
	return out
}

// FuenteTienePropiedad indica si una comunicación de esta fuente debe tener
// una Property asociada (viene de un portal) o no (whatsapp/ivr/viewphone).
func FuenteTienePropiedad(fuente string) bool {
	return !slices.Contains(fuentesSinPropiedad, fuente)
}

func (s *PropertyDBStore) GetProperty(portalId string, portal string) (*models.Propiedad, error) {
	property := models.Propiedad{}
	err := s.db.Get(&property, selectPropQ, portalId, portal)
	if err != nil {
		return nil, fmt.Errorf("the property with id %s doest not exists", portalId)
	}
	return &property, nil
}

func (s *PropertyDBStore) InsertProperty(p *models.Propiedad) error {
	res, err := s.db.NamedExec(insertPropQ, &p)
	if err != nil {
		return fmt.Errorf("error inserting property: %s", err.Error())
	}

	pId, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert id: %s", err.Error())
	}
	p.ID = models.NullInt32{Int32: int32(pId), Valid: true}
	return nil
}
