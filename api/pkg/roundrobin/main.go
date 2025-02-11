package roundrobin

import (
	"sync"
)

type RoundRobin[T comparable] struct {
	lock    sync.RWMutex
	current int

	participants []*T
}

func New[T comparable](participants []*T) *RoundRobin[T] {
    if (len(participants) == 0){
        panic("round-robin must have at least 1 participant")
    }

	r := RoundRobin[T]{
		current:  0,
        participants: participants,
	}

	return &r
}

/*
Insert a new participant into the list
*/
func (r *RoundRobin[T]) Add(p *T) {
	r.lock.RLock()
	defer r.lock.RUnlock()

    r.participants = append(r.participants, p)
}

func (r *RoundRobin[T]) Contains(participant *T) bool {
    for _, p := range r.participants {
        if *participant == *p {
            return true
        }
    }
    return false
}

func (r *RoundRobin[T]) Reasign(participants []*T) {
	r.lock.RLock()
	defer r.lock.RUnlock()

    r.participants = participants 
    r.Restart()
}

func (r *RoundRobin[T]) Restart() {
	r.lock.RLock()
	defer r.lock.RUnlock()

    r.current = 0
}

func (r *RoundRobin[T]) Next() *T {
	r.lock.RLock()
    defer r.lock.RUnlock()

	r.current += 1
	r.current %= len(r.participants)
	return r.participants[r.current]
}
