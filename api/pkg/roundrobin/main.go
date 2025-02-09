package roundrobin

import (
	"sync"
)

type RoundRobin[T any] struct {
	lock    sync.RWMutex
	current int

	participants []T
}

func New[T any](participants []T) *RoundRobin[T] {
    if (len(participants) == 0){
        panic("round-robin must have at least 1 participant")
    }

	r := RoundRobin[T]{
		current:  0,
		participants: participants,
	}

	return &r
}

func (r *RoundRobin[T]) Reasign(participants []T) {
	r.lock.RLock()
    r.participants = participants 
    r.Restart()
	r.lock.RUnlock()
}

func (r *RoundRobin[T]) Restart() {
	r.lock.RLock()
    r.current = 0
	r.lock.RUnlock()
}

func (r *RoundRobin[T]) Next() T {
	r.lock.RLock()
    defer r.lock.RUnlock()
	r.current += 1
	r.current %= len(r.participants)
	return r.participants[r.current]
}
