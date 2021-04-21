package timer

import (
	"math/rand"
	"time"

	"github.com/kitengo/raft/internal/rconfig"
)

type RaftTimer interface {
	SetDeadline(currentTime time.Time)
	GetDeadline() time.Time
	SetIdleTimeout()
	GetIdleTimeout() time.Time
}

func NewRaftTimer() RaftTimer  {
	deadlineTick := time.Now()
	return &raftTimer{
		deadlineTick: deadlineTick,
	}
}

type raftTimer struct{
	deadlineTick time.Time
	idleTick time.Time
}

func (rt *raftTimer) SetDeadline(currentTime time.Time) {
	rand.Seed(time.Now().UnixNano())
	rDelta := rand.Intn(rconfig.MaxPacketDelayMs - rconfig.MinPacketDelayMs) + rconfig.MinPacketDelayMs
	rt.deadlineTick = currentTime.Add(time.Duration(rDelta) * time.Millisecond)
}

func (rt *raftTimer) GetDeadline() time.Time {
	return rt.deadlineTick
}

func (rt *raftTimer) SetIdleTimeout() {
	rt.deadlineTick = time.Now().Add(rconfig.IdleTimeout)
}

func (rt *raftTimer) GetIdleTimeout() time.Time {
	return rt.deadlineTick
}


