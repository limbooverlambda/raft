package actors

import (
	"context"
	"fmt"
	"log"
	"time"

	raftlog "kitengo/raft/log"
	"kitengo/raft/rconfig"
	"kitengo/raft/timer"
	"kitengo/raft/transport"
	"kitengo/raft/voter"
)

type Follower interface {
	Run() FollowerCompletionChan
}

type FollowerTrigger struct{}

type FollowerCompletionChan <- chan struct{}

func NewFollower(ctx context.Context,
	followerTrigger <-chan FollowerTrigger,
	candidateTrigger chan<- CandidateTrigger,
	transport transport.Transport,
	raftTimer timer.RaftTimer,
	raftLog raftlog.RaftLog,
	raftVoter voter.RaftVoter,
) Follower {
	return &follower{
		context:          ctx,
		followerTrigger:  followerTrigger,
		candidateTrigger: candidateTrigger,
		transport:        transport,
		raftTimer:        raftTimer,
		raftLog:          raftLog,
		raftVoter:        raftVoter,
	}
}

type follower struct {
	context          context.Context
	followerTrigger  <-chan FollowerTrigger
	candidateTrigger chan<- CandidateTrigger
	transport        transport.Transport
	raftTimer        timer.RaftTimer
	raftLog          raftlog.RaftLog
	raftVoter        voter.RaftVoter
}

func (f *follower) Run() FollowerCompletionChan {
	followerCompletionChan := make(chan struct{})
	defer close(followerCompletionChan)
	go func() {
		fmt.Println("Follower running..")
		ticker := time.NewTicker(rconfig.PollDuration * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			select {
			case <-f.context.Done():
				log.Println("Cancelling Follower")
				return
			case <-f.followerTrigger:
				log.Println("Starting the follower")
				f.startFollower()
			default:
				log.Println("Follower ticking...")
			}
		}
	}()
	return followerCompletionChan
}

func (f *follower) startFollower() {
	ticker := time.NewTicker(rconfig.PollDuration * time.Second)
	defer ticker.Stop()
	for tick := range ticker.C {
		select {
		case req := <-f.transport.GetRequestChan():
			{
				//Process request
				f.processRequest(req)
				f.resetDeadline(tick)
			}
		default:
			{
				if f.exceedsDeadline(tick) {
					//transition to candidate
					f.transitionToCandidate()
					return
				}
			}
		}
	}
}

func (f *follower) processRequest(request transport.Request) {
	switch f.transport.GetRequestType(request) {
	case transport.AppendEntry:
		f.appendEntry(request)
	case transport.VoteRequest:
		f.processVote(request)
	}
}

func (f *follower) exceedsDeadline(tick time.Time) bool {
	return tick.After(f.raftTimer.GetDeadline())
}

func (f *follower) resetDeadline(tick time.Time) {
	//Resetting the deadline
	f.raftTimer.SetDeadline(tick)
}

func (f *follower) transitionToCandidate() {
	f.candidateTrigger <- CandidateTrigger{}
}

func (f *follower) appendEntry(request transport.Request) {
	//Logic for appending entry
	f.raftLog.AppendEntry(request.Payload)
}

func (f *follower) processVote(request transport.Request) {
	//Logic for processing vote
	f.raftVoter.ProcessVote(request.Payload)
}
