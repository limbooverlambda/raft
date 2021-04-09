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
		log.Println("Candidate running...")
		for range ticker.C {
			select {
			case <-f.context.Done():
				fmt.Println("Cancelling Candidate")
				return
			case <-f.candidateTrigger:
				fmt.Println("Candidate Triggered")
				sc := f.startCandidate()
				select {
				case <-sc:
					{
						return
					}
				case <-f.context.Done():
					{
						log.Println("Cancelling the candidate")
					}
				}
			default:
				fmt.Println("Candidate ticking...")
			}
		}
	}()
	return cChan
}

type candidateCompletionChan <-chan struct{}

func (f *candidate) startCandidate() candidateCompletionChan {
	cCompletionChan := make(chan struct{})
	go func() {
		defer close(cCompletionChan)
		eChan := f.startElection(f.context)
		for {
			select {
			case <-f.context.Done():
				return
			case <-eChan:
				return
			case appendEntryReq := <-f.raftAppendEntry.AppendEntryReqChan():
				{
					log.Println("Got an append entry request")
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
	}()
	return cCompletionChan
}

type electionCompletionChan <-chan struct{}

func (f *candidate) startElection(ctx context.Context) electionCompletionChan {
	ecChan := make(chan struct{})
	go func() {
		ticker := time.NewTicker(rconfig.PollDuration * time.Second)
		defer ticker.Stop()
		f.raftTimer.SetDeadline(time.Now())
		voteStatusChan := f.callElection()
		for tick := range ticker.C {
			select {
			case <-ctx.Done():
				log.Println("Cancelling election")
				ecChan <- struct{}{}
				return
			case voteStatus := <-voteStatusChan:
				{
					log.Printf("Got voteStatus %v\n", voteStatus)
					switch voteStatus {
					case voter.Leader:
						f.transitionToLeader(ctx, ecChan)
						return
					default:
						//Split vote or loss: Do nothing, re-election will be triggered
					}
				}
			default:
				if tick.After(f.raftTimer.GetDeadline()) {
					log.Println("Deadline exceeded, restarting the election")
					f.raftTimer.SetDeadline(tick)
					voteStatusChan = f.callElection()
				}
			}
		}
	}()
	return ecChan
}

func (f *candidate) transitionToLeader(ctxt context.Context, ecChan chan struct{}) {
	go func() {
		defer func() {
			ecChan <- struct{}{}
		}()
		select {
		case <-ctxt.Done():
			{
				log.Println("Cancelling transitioning to leader")
				return
			}
		case f.leaderTrigger <- LeaderTrigger{}:
			log.Println("Transitioned to leader")
			return
		}
	}()
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
