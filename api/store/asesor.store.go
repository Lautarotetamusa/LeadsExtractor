package store

import (
	"fmt"
	"leadsextractor/models"

	"github.com/jmoiron/sqlx"
)

type AsesorStorer interface {
    GetAll() (*[]models.Asesor, error)
    GetOne(string) (*models.Asesor, error)
    Insert(*models.Asesor) (*models.Asesor, error)
    Update(*models.Asesor, string) (*models.Asesor, error)
}

type AsesorMysqlStorage struct{
    Db *sqlx.DB
}

func (s *AsesorMysqlStorage) GetAll() (*[]models.Asesor, error){
    asesores := []models.Asesor{}
    if err := s.Db.Select(&asesores, "SELECT * FROM Asesor"); err != nil{
        return nil, err
    }
    return &asesores, nil 
}

func (s *AsesorMysqlStorage) GetOne(phone string) (*models.Asesor, error){
    asesor := models.Asesor{}
    if err := s.Db.Get(&asesor, "SELECT * FROM Asesor WHERE phone=?", phone); err != nil{ 
        return nil, err
    }
    return &asesor, nil 
}

func (s *AsesorMysqlStorage) Insert(asesor *models.Asesor) (*models.Asesor, error){
    query := "INSERT INTO Asesor (name, phone, active) VALUES (:name, :phone, :active)"
    if _, err := s.Db.NamedExec(query, asesor); err != nil {
        fmt.Printf("%v", err)
        return nil, err
    }
    return asesor, nil 
}

func (s *AsesorMysqlStorage) Update(asesor *models.Asesor, phone string) (*models.Asesor, error){
    query := "UPDATE Asesor SET name=:name, active=:active WHERE phone=:phone"
    if _, err := s.Db.NamedExec(query, asesor); err != nil {
        fmt.Printf("%v", err)
        return nil, err
    }
    return asesor, nil 
}
