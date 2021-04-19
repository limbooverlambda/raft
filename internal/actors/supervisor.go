package actors

import (
	"log"

	rstate "github.com/kitengo/raft/internal/state"
)

type RaftSupervisor interface {
	//Start the Supervisor routine
	Start()
}

func NewRaftSupervisor() RaftSupervisor {
	return &raftSupervisor{
		raftState:         rstate.NewRaftState(),
		followerProvider:  followerProvider{},
		leaderProvider:    leaderProvider{},
		candidateProvider: candidateProvider{},
	}
}

type raftSupervisor struct {
	raftState         rstate.RaftState
	followerProvider  FollowerProvider
	leaderProvider    LeaderProvider
	candidateProvider CandidateProvider
}

func (rs *raftSupervisor) Start() {
	loop:
	for {
		select {
		case state := <-rs.raftState.GetStateChan():
			switch state {
			case rstate.FollowerState:
				rs.runFollower()
			case rstate.CandidateState:
				rs.runCandidate()
			case rstate.LeaderState:
				rs.runLeader()
			case rstate.ShutdownState:
				rs.shutdown()
				break loop
			}
		}
	}
}

func (rs *raftSupervisor) shutdown() {
	log.Println("Shutdown, relooping")
}

func (rs *raftSupervisor) runLeader() {
	leader := rs.leaderProvider.Provide(rs.raftState)
	leader.Run()
}

func (rs *raftSupervisor) runCandidate() {
	candidate := rs.candidateProvider.Provide(rs.raftState)
	candidate.Run()
}

func (rs *raftSupervisor) runFollower() {
	follower := rs.followerProvider.Provide(rs.raftState)
	follower.Run()
}
