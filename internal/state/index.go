package state

import (
	"sync"
)

type RaftIndex struct {
	sync.RWMutex
	commitOffset int64
	applyOffset int64
}

func NewRaftIndex() *RaftIndex {
	return &RaftIndex{}
}

func (ri *RaftIndex) SetCommitOffset(offset int64) {
	ri.Lock()
	defer ri.Unlock()
	ri.commitOffset = offset
}

func (ri *RaftIndex) SetApplyOffset(offset int64)  {
	ri.Lock()
	defer ri.Unlock()
	ri.applyOffset = offset
}

func (ri *RaftIndex) GetCommitOffset() int64  {
	ri.RLock()
	defer ri.RUnlock()
	return ri.commitOffset
}

func (ri *RaftIndex) GetApplyOffset() int64  {
	ri.RLock()
	defer ri.RUnlock()
	return ri.applyOffset
}

