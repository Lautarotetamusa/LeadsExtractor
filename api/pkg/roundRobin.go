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
    asesores := []models.Asesor{}

    err := db.Select(&asesores, "SELECT * FROM Asesor WHERE active=?", true)
    if err != nil{
        log.Fatal("No se pudo obtener la lista de asesores\n", err)
    }

    return &RoundRobin{
        current: 0,
        asesores: asesores,
    }
}

func (r *RoundRobin) SetAsesores(db *sqlx.DB) {
    asesores := []models.Asesor{}

    err := db.Select(&asesores, "SELECT * FROM Asesor WHERE active=?", true)
    if err != nil{
        log.Fatal("No se pudo obtener la lista de asesores\n", err)
    }

    r.lock.RLock()
    r.asesores = asesores
    r.lock.RUnlock()
}

func (r *RoundRobin) next() models.Asesor{
    r.lock.RLock()
    r.current += 1
    r.current %= len(r.asesores)
    r.lock.RUnlock()
    return r.asesores[r.current]
}

