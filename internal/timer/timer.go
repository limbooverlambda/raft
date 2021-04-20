package timer

import (
	"time"
)

type RaftTimer interface {
	SetDeadline(currentTime time.Time)
	GetDeadline() time.Time
	SetIdleTimeout()
	GetIdleTimeout() time.Time
}

func NewRaftTimer() RaftTimer  {
	return raftTimer{}
}

type raftTimer struct{}

func (raftTimer) SetDeadline(currentTime time.Time) {
	panic("implement me")
}

func (raftTimer) GetDeadline() time.Time {
	panic("implement me")
}

func (raftTimer) SetIdleTimeout() {
	panic("implement me")
}

func (raftTimer) GetIdleTimeout() time.Time {
	panic("implement me")
}


