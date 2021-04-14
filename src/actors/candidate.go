package actors

import (
	"log"

	"kitengo/raft/state"
)

type Candidate interface {
	Run()
}

type candidate struct{
	state.RaftState
}

func (c *candidate) Run() {
	log.Println("Candidate")
	c.SetState(state.LeaderState)
}

type CandidateProvider interface {
	Provide(raftState state.RaftState) Candidate
}

type candidateProvider struct{}

func (candidateProvider) Provide(raftState state.RaftState) Candidate {
	return &candidate{raftState}
}
