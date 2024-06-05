package store

import (
	"leadsextractor/models"
	"log"
	"sync"
)

type RoundRobin struct {
	lock    sync.RWMutex
	current int

	asesores []models.Asesor
}

func NewRoundRobin(asesores []models.Asesor) *RoundRobin {
    if (len(asesores) == 0){
        log.Fatal("No se puede crear un round robin sin asesores")
    }

	r := RoundRobin{
		current:  0,
		asesores: asesores,
	}

	return &r
}

func (r *RoundRobin) SetAsesores(asesores []models.Asesor) {
	r.lock.RLock()
    r.asesores = asesores

	r.lock.RUnlock()
}

func (r *RoundRobin) Next() models.Asesor {
	r.lock.RLock()
	r.current += 1
	r.current %= len(r.asesores)
	r.lock.RUnlock()
	return r.asesores[r.current]
}
