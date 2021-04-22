package term

import (
	"sync"
)

type RaftTerm interface {
	GetTerm() int64
	IncrementTerm() int64
}

func NewRaftTerm() RaftTerm {
	return &raftTerm{}
}

type raftTerm struct{
	sync.RWMutex
	term int64
}

func (rt *raftTerm) GetTerm() int64 {
	defer rt.RUnlock()
	rt.RLock()
	return rt.term
}

func (rt *raftTerm) IncrementTerm() int64 {
	defer rt.Unlock()
	rt.Lock()
	rt.term += 1
	return rt.term
}
