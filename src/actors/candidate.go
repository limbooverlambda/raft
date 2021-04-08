package actors

import (
	"context"
	"fmt"
	"log"
	"time"

	"kitengo/raft/rconfig"
	"kitengo/raft/rpc"
	"kitengo/raft/term"
	"kitengo/raft/timer"
	"kitengo/raft/voter"
)

//TODO: Write unit tests for the Candidate scaffolding
//TODO: Add failure conditions for the unit test.
//TODO: Figure out how the transport will be plugged into the server

type CandidateTrigger struct{}

type CandidateCompletionChan <-chan struct{}

type Candidate interface {
	Run() CandidateCompletionChan
}

func NewCandidate(ctx context.Context,
	leaderTrigger chan<- LeaderTrigger,
	candidateTrigger <-chan CandidateTrigger) Candidate {
	return &candidate{
		context:          ctx,
		leaderTrigger:    leaderTrigger,
		candidateTrigger: candidateTrigger,
	}
}

type candidate struct {
	context          context.Context
	leaderTrigger    chan<- LeaderTrigger
	candidateTrigger <-chan CandidateTrigger
	followerTrigger  chan<- FollowerTrigger
	raftTimer        timer.RaftTimer
	raftTerm         term.RaftTerm
	raftVoter        voter.RaftVoter
	raftAppendEntry  rpc.RaftAppendEntry
}

func (f *candidate) Run() CandidateCompletionChan {
	cChan := make(chan struct{})
	go func() {
		defer close(cChan)
		ticker := time.NewTicker(rconfig.PollDuration * time.Second)
		defer ticker.Stop()
		fmt.Println("Candidate running...")
		for range ticker.C {
			fmt.Println("Candidate ticking")
			select {
			case <-f.context.Done():
				fmt.Println("Cancelling Candidate")
				return
			case <-f.candidateTrigger:
				fmt.Println("Candidate Triggered")
				f.startCandidate()
			default:
				fmt.Println("Candidate ticking...")
			}
		}
	}()
	return cChan
}

func (f *candidate) startCandidate() {
	eChan := f.startElection(f.context)
	for {
		select {
		case <- f.context.Done():
			return
		case <-eChan:
			return
		case appendEntryReq := <-f.raftAppendEntry.AppendEntryReqChan():
			{
				resp := f.raftAppendEntry.Process(appendEntryReq)
				f.raftAppendEntry.AppendEntryRespChan() <- resp
				if resp.Success {
					//Append succeeded, flip back to being a follower
					f.followerTrigger <- FollowerTrigger{}
					return
				}
			}
		}
	}
}

func (f *candidate) startElection(ctx context.Context) <-chan struct{} {
	electionCompletionChan := make(chan struct{})
	go func() {
		ticker := time.NewTicker(rconfig.PollDuration * time.Second)
		defer ticker.Stop()
		f.raftTimer.SetDeadline(time.Now())
		voteStatusChan := f.callElection()
		for tick := range ticker.C {
			select {
			case <-ctx.Done():
				log.Println("Cancelling election")
				electionCompletionChan <- struct{}{}
				return
			case voteStatus := <-voteStatusChan:
				{
					switch voteStatus {
					case voter.Leader:
						f.leaderTrigger <- LeaderTrigger{}
						electionCompletionChan <- struct{}{}
						return
					default:
						//Split vote or loss: Do nothing, re-election will be triggered
					}
				}
			default:
				if f.raftTimer.GetDeadline().After(tick) {
					f.raftTimer.SetDeadline(tick)
					voteStatusChan = f.callElection()
				}
			}
		}
	}()
	return electionCompletionChan
}

/*
• Increment currentTerm
• Vote for self
• Reset election timer
• Send RequestVote RPCs to all other servers
*/
func (f *candidate) callElection() <-chan voter.VoteStatus {
	term := f.raftTerm.IncrementTerm()
	return f.raftVoter.RequestVote(term)
}
