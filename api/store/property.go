package store

import (
	"fmt"
	"leadsextractor/models"
	"time"

	"github.com/jmoiron/sqlx"
)

// CREATE TABLE IF NOT EXISTS PortalProp(
//     id INT NOT NULL AUTO_INCREMENT,
//
//     title VARCHAR(256) NOT NULL,
//     price VARCHAR(32) NOT NULL,
//     currency CHAR(5) NOT NULL,
//     description VARCHAR(512),
//     type VARCHAR(32) DEFAULT NULL,
//     antiquity INT NOT NULL,
//     parkinglots INT DEFAULT NULL,
//     bathrooms INT DEFAULT NULL,
//     half_bathrooms INT DEFAULT NULL,
//     rooms INT DEFAULT NULL,
//     operation_type ENUM("sale", "rent") NOT NULL,
//     m2_total INT,
//     m2_covered INT,
//     video_url VARCHAR(256) DEFAULT NULL,
//     virtual_route VARCHAR(256) DEFAULT NULL,
//
//     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
//
//     PRIMARY KEY(id)
// );

type PortalProp struct {
	ID            int64          `json:"id" db:"id"`

    Title         string            `json:"title" db:"title" validate:"required"`
    Price         string            `json:"price" db:"price" validate:"required,numeric,gte=0"`
    Currency      string            `json:"currency" db:"currency" validate:"required"`
    Description   string            `json:"description" db:"description" validate:"required"`
	Type          models.NullString `json:"type" db:"type"`
    Antiquity     int               `json:"antiquity" db:"antiquity" validate:"required"`
	ParkingLots   models.NullInt16  `json:"parking_lots" db:"parkinglots"`
	Bathrooms     models.NullInt16  `json:"bathrooms" db:"bathrooms"`
	HalfBathrooms models.NullInt16  `json:"half_bathrooms" db:"half_bathrooms"`
	Rooms         models.NullInt16  `json:"rooms" db:"rooms"`
    OperationType string            `json:"operation_type" db:"operation_type" validate:"required,oneof=sale rent"`
	M2Total       models.NullInt16  `json:"m2_total" db:"m2_total"`
	M2Covered     models.NullInt16  `json:"m2_covered" db:"m2_covered"`
	VideoURL      models.NullString `json:"video_url" db:"video_url"`
	VirtualRoute  models.NullString `json:"virtual_route" db:"virtual_route"`

    State	     string             `json:"state" db:"state" validate:"required"`
    Municipality string             `json:"municipality" db:"municipality" validate:"required"`
    Colony	     string             `json:"colony" db:"colony" validate:"required"`
    Neighborhood models.NullString  `json:"neighborhood" db:"neighborhood"`
    Street	     string             `json:"street" db:"street" validate:"required"`
    Number	     string             `json:"number" db:"number" validate:"required"`
    ZipCode	 string                 `json:"zip_code" db:"zip_code" validate:"required"`

	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

    Images []PropertyImage `json:"images"`
}

type PropertyImage struct {
    ID         int64    `json:"id" db:"id"`
    PropertyID int64    `json:"-" db:"property_id"`
    Url        string   `json:"url" db:"url"` 
}

const (
    createPropertyQ = `
        INSERT INTO PortalProp 
                (title, price, currency, description, type, antiquity, parkinglots, bathrooms, half_bathrooms, rooms, 
                operation_type, m2_total, m2_covered, video_url, virtual_route,
                state, municipality, colony, neighborhood, street, number, zip_code
            ) 
        VALUES (:title, :price, :currency, :description, :type, :antiquity, :parkinglots, :bathrooms, :half_bathrooms, :rooms, 
                :operation_type, :m2_total, :m2_covered, :video_url, :virtual_route,
                :state, :municipality, :colony, :neighborhood, :street, :number, :zip_code
            )`

	updatePropertyQ = `
		UPDATE PortalProp 
		SET title = :title, price = :price, currency = :currency, description = :description, type = :type,
			antiquity = :antiquity, parkinglots = :parkinglots, bathrooms = :bathrooms, 
			half_bathrooms = :half_bathrooms, rooms = :rooms, operation_type = :operation_type, 
			m2_total = :m2_total, m2_covered = :m2_covered, video_url = :video_url, virtual_route = :virtual_route,
            state = :state, municipality = :municipality, colony = :colony, neighborhood = :neighborhood, 
            street = :street, number = :number, zip_code = :zip_code 
		WHERE id = :id`
)

type PropertyPortalStore interface {
	GetAll() ([]*PortalProp, error)
	GetOne(int64) (*PortalProp, error)
    GetImages(*PortalProp) error
	Insert(*PortalProp) error
	InsertImages(int64, []PropertyImage) error
	Update(*PortalProp) error
    DeleteImage(propId, imageId int64) error
}

type propertyPortalStore struct {
	db *sqlx.DB
}

func NewPropertyPortalStore(db *sqlx.DB) PropertyPortalStore {
	return &propertyPortalStore{db: db}
}

func (s *propertyPortalStore) GetAll() ([]*PortalProp, error) {
    // i do it this way because if not are any props must return [] and not null
    properties := make([]*PortalProp, 0)

	query := `SELECT * FROM PortalProp`
	err := s.db.Select(&properties, query)
	if err != nil {
		return nil, fmt.Errorf("error al obtener todas las propiedades: %w", err)
	}

	return properties, nil
}

func (s *propertyPortalStore) GetOne(id int64) (*PortalProp, error) {
	var property PortalProp
	query := `SELECT * FROM PortalProp WHERE id = ?`
	err := s.db.Get(&property, query, id)
	if err != nil {
		return nil, SQLNotFound(err, "property not found")
	}
	return &property, nil
}

func (s *propertyPortalStore) GetImages(prop *PortalProp) error {
	imagesQuery := `SELECT id, url FROM PropertyImages WHERE property_id = ?`
    err := s.db.Select(&prop.Images, imagesQuery, prop.ID)
	if err != nil {
        return SQLNotFound(err, fmt.Sprintf("error obtaining images for the property %d: %w", prop.ID, err))
	}

    return nil
}

func (s *propertyPortalStore) Insert(prop *PortalProp) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
    
	res, err := tx.NamedExec(createPropertyQ, prop)
	if err != nil {
        tx.Rollback()
        return err
	}

    prop.ID, err = res.LastInsertId()
	if err != nil {
        tx.Rollback()
        return err
	}

    if len(prop.Images) > 0 {
        if err := insertImages(tx, prop.ID, prop.Images); err != nil {
            tx.Rollback()
            return err
        }
    }

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error commiting the transaction: %w", err)
	}

    // get the created_at, updated_at and id fields
    prop, err = s.GetOne(prop.ID)
	return err 
}

func (s *propertyPortalStore) InsertImages(propId int64, images []PropertyImage) error {
    return insertImages(s.db, propId, images)
}

func insertImages(ext sqlx.Ext, propId int64, images []PropertyImage) error {
    for i, _ := range images {
        images[i].PropertyID = propId
    }

    query := "INSERT INTO PropertyImages (property_id, url) VALUES (:property_id, :url)" 
    _, err := sqlx.NamedExec(ext, query, images)
	if err != nil {
        return SQLBadForeignKey(err, fmt.Sprintf("property with id %d does not exists", propId))
	}
    return nil
}

func (s *propertyPortalStore) DeleteImage(propId, imageId int64) error {
    query := "DELETE FROM PropertyImages WHERE property_id = ? AND id = ?"
    res, err := s.db.Exec(query, propId, imageId)

    affected, err := res.RowsAffected()
    if err != nil || affected == 0 {
        return NewErr("image does not exists", StoreNotFoundErr)
    }

    return err
}

func (s *propertyPortalStore) Update(prop *PortalProp) error {
	res, err := s.db.NamedExec(updatePropertyQ, prop)
    affected, err := res.RowsAffected()
    if err != nil || affected == 0 {
        return NewErr("property doest not exists", StoreNotFoundErr)
    }
	if err != nil {
		return SQLNotFound(err, fmt.Sprintf("error updating the property with id %d: %w", prop.ID, err))
	}

    return nil
}
