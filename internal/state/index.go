package state

import (
	"sync"
)

type RaftIndex struct {
	sync.RWMutex
	commitOffset uint64
	applyOffset  uint64
}

func NewRaftIndex() *RaftIndex {
	return &RaftIndex{}
}

func (ri *RaftIndex) SetCommitOffset(offset uint64) {
	ri.Lock()
	defer ri.Unlock()
	ri.commitOffset = offset
}

func (ri *RaftIndex) SetApplyOffset(offset uint64) {
	ri.Lock()
	defer ri.Unlock()
	ri.applyOffset = offset
}

func (ri *RaftIndex) GetCommitOffset() uint64 {
	ri.RLock()
	defer ri.RUnlock()
	return ri.commitOffset
}

func (ri *RaftIndex) GetApplyOffset() uint64 {
	ri.RLock()
	defer ri.RUnlock()
	return ri.applyOffset
}
