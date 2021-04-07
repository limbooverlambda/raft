package actors

import (
	"context"
	"fmt"
	"time"

	"kitengo/raft/rconfig"
)

type CandidateTrigger struct{}

type Candidate interface {
	Run()
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
}

func (f *candidate) Run() {
	go func() {
		ticker := time.NewTicker(rconfig.PollDuration * time.Second)
		defer ticker.Stop()
		fmt.Println("Candidate running...")
		for range ticker.C {
			fmt.Println("Candidate ticking")
			select {
			case <-f.context.Done():
				fmt.Println("Cancelling Candidate")
				return
			case <- f.candidateTrigger:
				fmt.Println("Candidate Triggered")
				f.leaderTrigger <- LeaderTrigger{}
			default:
				fmt.Println("Candidate ticking...")
			}
		}
	}()
}
