package actors

import (
	"context"
	"log"

	svclocator "github.com/kitengo/raft/internal/locator"
	rstate "github.com/kitengo/raft/internal/state"
)

type RaftSupervisor interface {
	//Start the Supervisor routine
	Start(ctx context.Context)
}

func NewRaftSupervisor(locator svclocator.ServiceLocator) RaftSupervisor {
	return &raftSupervisor{
		raftState: locator.GetRaftState(),
		followerProvider: &followerProvider{
			locator,
		},
		leaderProvider: &leaderProvider{
			locator,
		},
		candidateProvider: &candidateProvider{locator},
	}
}

type raftSupervisor struct {
	raftState         rstate.RaftState
	followerProvider  FollowerProvider
	leaderProvider    LeaderProvider
	candidateProvider CandidateProvider
}

func (rs *raftSupervisor) Start(ctx context.Context) {
loop:
	for {
		select {
		case state := <-rs.raftState.GetStateChan():
			switch state {
			case rstate.FollowerState:
				rs.runFollower(ctx)
			case rstate.CandidateState:
				rs.runCandidate(ctx)
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
	leader := rs.leaderProvider.Provide()
	leader.Run()
}

func (rs *raftSupervisor) runCandidate(ctx context.Context) {
	candidate := rs.candidateProvider.Provide()
	candidate.Run(ctx)
}

func (rs *raftSupervisor) runFollower(ctx context.Context) {
	follower := rs.followerProvider.Provide()
	follower.Run(ctx)
}
