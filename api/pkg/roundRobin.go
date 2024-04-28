package pkg

import (
	"leadsextractor/models"
	"log"
	"sync"

	"github.com/jmoiron/sqlx"
)

type RoundRobin struct{
    lock        sync.RWMutex
    current     int

    asesores    []models.Asesor
}

func NewRoundRobin(db *sqlx.DB) *RoundRobin{
    r := RoundRobin{
        current: 0,
        asesores: []models.Asesor{},
    }
    r.SetAsesores(db)
    return &r
}

func (r *RoundRobin) SetAsesores(db *sqlx.DB) {
    r.lock.RLock()

    err := db.Select(&r.asesores, "SELECT * FROM Asesor WHERE active=?", true)
    if err != nil{
        log.Fatal("No se pudo obtener la lista de asesores\n", err)
    }

    r.lock.RUnlock()
}

func (r *RoundRobin) next() models.Asesor{
    r.lock.RLock()
    r.current += 1
    r.current %= len(r.asesores)
    r.lock.RUnlock()
    return r.asesores[r.current]
}

