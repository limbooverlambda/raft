package actors

import (
	"log"

	"kitengo/raft/state"
)

type Leader interface {
	Run()
}

type leader struct{
	state.RaftState
}

func (l *leader) Run() {
	log.Println("Leader")
	l.SetState(state.ShutdownState)
}

type LeaderProvider interface {
	Provide(raftState state.RaftState) Leader
}

type leaderProvider struct{}

func (leaderProvider) Provide(raftState state.RaftState) Leader {
	return &leader{raftState}
}


