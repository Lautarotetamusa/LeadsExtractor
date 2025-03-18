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
//     description VARCHAR(512) NOT NULL, 
//     type VARCHAR(32) NOT NULL,
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
//     /* Ubication fields */
//     address VARCHAR(256) NOT NULL, /*Full GOOGLE valid address*/
//     lat FLOAT NOT NULL,
//     lng FLOAT NOT NULL,
//     PRIMARY KEY(id)
// )

type Location struct {
    Lat    float32 `json:"lat" db:"lat" validate:"required"`
    Lng    float32 `json:"lng" db:"lng" validate:"required"`
}

type Ubication struct {
    Address     string   `json:"address" csv:"address" db:"address" validate:"required"`
    Location    Location `json:"location"`
}

type PortalProp struct {
	ID            int64          `json:"id" db:"id"`

    // This fields are generated in case of creation with csv
    Title         string            `json:"title" db:"title" validate:"required"`
    Description   string            `json:"description" db:"description" validate:"required"`

    Price         string            `json:"price" db:"price" csv:"price" validate:"required,numeric,gte=0"`
    Currency      string            `json:"currency" db:"currency" csv:"currency" validate:"required"`
    Type          string            `json:"type" db:"type" csv:"type" validate:"required,oneof=house apartment"`
    Antiquity     int               `json:"antiquity" csv:"antiquity" db:"antiquity" validate:"required"`
    ParkingLots   int               `json:"parking_lots" csv:"parking_lots" db:"parkinglots"`
    Bathrooms     int               `json:"bathrooms" csv:"bathrooms" db:"bathrooms"`
    HalfBathrooms int               `json:"half_bathrooms" csv:"half_bathrooms" db:"half_bathrooms"`
    Rooms         int               `json:"rooms" csv:"rooms" db:"rooms"`
    OperationType string            `json:"operation_type" csv:"operation_type" db:"operation_type" validate:"required,oneof=sell rent"`
    M2Total       models.NullInt16  `json:"m2_total" csv:"m2_total" db:"m2_total"`
    M2Covered     models.NullInt16  `json:"m2_covered" csv:"m2_covered" db:"m2_covered"`
    VideoURL      models.NullString `json:"video_url" db:"video_url"`
	VirtualRoute  models.NullString `json:"virtual_route" db:"virtual_route"`

    Ubication   Ubication   `json:"ubication"`

	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

    Images []PropertyImage `json:"images,omitempty"`
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
                address, lat, lng
            ) 
        VALUES (:title, :price, :currency, :description, :type, :antiquity, :parkinglots, :bathrooms, :half_bathrooms, :rooms, 
                :operation_type, :m2_total, :m2_covered, :video_url, :virtual_route,
                :ubication.address, :ubication.location.lat, :ubication.location.lng
            )`

	updatePropertyQ = `
		UPDATE PortalProp 
		SET title = :title, price = :price, currency = :currency, description = :description, type = :type,
			antiquity = :antiquity, parkinglots = :parkinglots, bathrooms = :bathrooms, 
			half_bathrooms = :half_bathrooms, rooms = :rooms, operation_type = :operation_type, 
			m2_total = :m2_total, m2_covered = :m2_covered, video_url = :video_url, virtual_route = :virtual_route,
            address = :ubication.address, lat = :ubication.location.lat, lng = :ubication.location.lng
		WHERE id = :id`

    selectPropertyQ = `
        SELECT id, title, price, currency, description, type, antiquity, parkinglots, bathrooms, half_bathrooms,
               rooms, operation_type, m2_total, m2_covered, video_url, virtual_route, created_at, updated_at,
               address as "ubication.address",
               lat as "ubication.location.lat",
               lng as "ubication.location.lng"
        FROM PortalProp`
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

	err := s.db.Select(&properties, selectPropertyQ)
	if err != nil {
		return nil, fmt.Errorf("error al obtener todas las propiedades: %w", err)
	}

	return properties, nil
}

func (s *propertyPortalStore) GetOne(id int64) (*PortalProp, error) {
	var property PortalProp
	query := selectPropertyQ + " WHERE id = ?"
	err := s.db.Get(&property, query, id)
	if err != nil {
		return nil, SQLNotFound(err, "property not found")
	}
	return &property, nil
}

func (s *propertyPortalStore) GetImages(prop *PortalProp) error {
	imagesQuery := "SELECT id, url FROM PropertyImages WHERE property_id = ?"
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
	return err 
}

func (s *propertyPortalStore) InsertImages(propId int64, images []PropertyImage) error {
    return insertImages(s.db, propId, images)
}

// The properties ID are updated
func insertImages(ext sqlx.Ext, propId int64, images []PropertyImage) error {
    for i, _ := range images {
        images[i].PropertyID = propId
    }

    query := "INSERT INTO PropertyImages (property_id, url) VALUES (:property_id, :url)" 
    res, err := sqlx.NamedExec(ext, query, images)
	if err != nil {
        return SQLBadForeignKey(err, fmt.Sprintf("property with id %d does not exists", propId))
	}

    // i dont know why, but actually the LastInsertId() gives us the id of the FIRST image.
    firstId, err := res.LastInsertId()
    if err != nil {
        return fmt.Errorf("error getting last inserting id %w", err)
    }

    for i, _ := range images {
        images[i].ID = firstId
        firstId++
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
	_, err := s.db.NamedExec(updatePropertyQ, prop)
	if err != nil {
		return fmt.Errorf("error updating the property with id %d: %w", prop.ID, err)
	}

    return nil
}
