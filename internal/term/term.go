package term

import (
	"log"
	"sync"
)

type RaftTerm interface {
	GetTerm() int64
	IncrementTerm() int64
}

func NewRaftTerm() RaftTerm {
	return &raftTerm{}
}

type raftTerm struct {
	sync.RWMutex
	term int64
}

func (rt *raftTerm) GetTerm() int64 {
	rt.RLock()
	defer rt.RUnlock()
	return rt.term
}

func (rt *raftTerm) IncrementTerm() int64 {
	rt.Lock()
	defer rt.Unlock()
	rt.term += 1
	log.Println("Incremented term to", rt.term)
	return rt.term
}
