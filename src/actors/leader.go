package actors

import (
	"context"
	"fmt"
	"time"

	"kitengo/raft/peers"
	"kitengo/raft/rconfig"
	"kitengo/raft/timer"
)

type Leader interface {
	Run()
}

//TODO: Work on the Leader scaffolding
//TODO: Unit tests for the leader scaffolding.
//TODO: Write comments for all the functions
//TODO: Make the unit tests faster
type LeaderTrigger struct{}

func NewLeader(ctx context.Context,
	leaderTrigger <-chan LeaderTrigger) Leader {
	return &leader{
		context:       ctx,
		leaderTrigger: leaderTrigger,
	}
}

type leader struct {
	context       context.Context
	peers         peers.RaftPeers
	raftTimer     timer.RaftTimer
	leaderTrigger <-chan LeaderTrigger
}

func (f *leader) Run() {
	go func() {
		ticker := time.NewTicker(rconfig.PollDuration * time.Second)
		defer ticker.Stop()
		fmt.Println("Leader running")
		for range ticker.C {
			fmt.Println("Leader ticking")
			select {
			case <-f.context.Done():
				fmt.Println("Cancelling Leader")
				return
			case <-f.leaderTrigger:
				fmt.Println("Leader triggered")
			default:
				fmt.Println("Leader ticking...")
			}
		}
	}()
}

/*
Upon election: send initial empty AppendEntries RPCs
(heartbeat) to each server; repeat during idle periods to
prevent election timeouts
• If command received from client: append entry to local log,
respond after entry applied to state machine
• If last log index ≥ nextIndex for a follower: send
AppendEntries RPC with log entries starting at nextIndex
• If successful: update nextIndex and matchIndex for
follower
• If AppendEntries fails because of log inconsistency:
decrement nextIndex and retry
• If there exists an N such that N > commitIndex, a majority
of matchIndex[i] ≥ N, and log[N].term == currentTerm:
set commitIndex = N
*/
func (f *leader) startLeader() {
	//Send heartbeats after idle time has expired
	//Process command when received from the command channel

}
