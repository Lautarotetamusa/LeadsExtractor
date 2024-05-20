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

func NewRoundRobin(s *Store) *RoundRobin {
	r := RoundRobin{
		current:  0,
		asesores: []models.Asesor{},
	}
	r.SetAsesores(s)
	return &r
}

func (r *RoundRobin) SetAsesores(s *Store) {
	r.lock.RLock()
	err := s.GetAllAsesores(&r.asesores)

	if err != nil {
		log.Fatal("No se pudo obtener la lista de asesores\n", err)
	}

	r.lock.RUnlock()
}

func (r *RoundRobin) Next() models.Asesor {
	r.lock.RLock()
	r.current += 1
	r.current %= len(r.asesores)
	r.lock.RUnlock()
	return r.asesores[r.current]
}
