package actors

import (
	"context"
	"testing"
	"time"
)

func Test_candidate_Run_WithCancellation(t *testing.T) {
	ctx := context.Background()
	cc, cancel := context.WithCancel(ctx)
	candidateTrigger := make(<-chan CandidateTrigger)
	leaderTrigger := make(chan<- LeaderTrigger)
	f := NewCandidate(cc,
		leaderTrigger,
		candidateTrigger,
	)
	c := f.Run()
	time.Sleep(3 * time.Second)
	cancel()
	<-c
}

