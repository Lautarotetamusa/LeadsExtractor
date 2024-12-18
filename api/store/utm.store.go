package store

import (
	"fmt"
	"leadsextractor/models"
	"strings"
)

func (s *Store) GetAllUtm(utms *[]models.UtmDefinition) error {
	if err := s.db.Select(utms, "SELECT * FROM Utm ORDER BY id DESC"); err != nil {
		return err
	}
	return nil
}

func (s *Store) GetOneUtm(id int) (*models.UtmDefinition, error) {
	utm := models.UtmDefinition{}
	if err := s.db.Get(&utm, "SELECT * FROM Utm WHERE id=?", id); err != nil {
		return nil, err
	}
	return &utm, nil
}

func (s *Store) GetOneUtmByCode(code string) (*models.UtmDefinition, error) {
	utm := models.UtmDefinition{}
	if err := s.db.Get(&utm, "SELECT * FROM Utm WHERE code=?", code); err != nil {
		return nil, err
	}
	return &utm, nil
}

func (s *Store) InsertUtm(utm *models.UtmDefinition) (int64, error) {
	query := `
    INSERT INTO Utm 
            ( code,  utm_source, utm_medium,  utm_campaign,  utm_ad,  utm_channel) 
    VALUES  (:code, :utm_source, :utm_medium, :utm_campaign, :utm_ad, :utm_channel)`

    res, err := s.db.NamedExec(query, utm)
    if err != nil {
        if strings.Contains(err.Error(), "Error 1062") {
            return 0, fmt.Errorf("ya existe un utm con codigo %s", utm.Code)
        }
        return 0, err
    }
    id, err := res.LastInsertId()
    if err != nil {
        return 0, err
    }
	return id, nil
}

func (s *Store) UpdateUtm(utm *models.UtmDefinition) error {
	query := `
    UPDATE Utm 
    SET utm_source = :utm_source,
        utm_medium = :utm_medium ,
        utm_campaign = :utm_campaign,
        utm_ad = :utm_ad,
        utm_channel = :utm_channel
    WHERE id=:id`

	if _, err := s.db.NamedExec(query, utm); err != nil {
		return err
	}
	return nil
}
