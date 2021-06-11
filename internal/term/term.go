package term

import (
	"log"
	"sync"
)

type RaftTerm interface {
	GetTerm() uint64
	IncrementTerm() uint64
	SetTerm(term uint64)
}

func NewRaftTerm() RaftTerm {
	return &raftTerm{}
}

type raftTerm struct {
	sync.RWMutex
	term uint64
}

func (rt *raftTerm) SetTerm(term uint64) {
	rt.Lock()
	defer rt.Unlock()
	rt.term = term
}

func (rt *raftTerm) GetTerm() uint64 {
	rt.RLock()
	defer rt.RUnlock()
	return rt.term
}

func (rt *raftTerm) IncrementTerm() uint64 {
	rt.Lock()
	defer rt.Unlock()
	rt.term += 1
	log.Println("Incremented term to", rt.term)
	return rt.term
}
