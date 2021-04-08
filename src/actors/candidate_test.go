package actors

import (
	"context"
	"reflect"
	"testing"
	"time"

	"kitengo/raft/voter"
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
	time.Sleep(1 * time.Second)
	cancel()
	<-c
}

func Test_candidate_startElection_withLeaderSelectedHappyPath(t *testing.T) {
	leaderTrigger := make(chan LeaderTrigger)
	f := instantiateDefaultCandidate(candidateStub{leaderTrigger: leaderTrigger})
	ctxt, cancel := context.WithCancel(context.Background())

	ec := f.startElection(ctxt)
	time.AfterFunc(2*time.Second, cancel)

	expectedLeaderTrigger := LeaderTrigger{}
	if <-leaderTrigger != expectedLeaderTrigger {
		t.Error("Expected completion signal but received something else")
	}
	<- ec
}

func Test_candidate_startElection_withLeaderSelectedWithoutReceiver(t *testing.T) {
	f := instantiateDefaultCandidate(candidateStub{})
	ctxt, cancel := context.WithCancel(context.Background())
	ec := f.startElection(ctxt)
	time.AfterFunc(1*time.Second, cancel)
	signal := <-ec
	if signal != struct{}{} {
		t.Error("Expected completion signal but received something else")
	}
}

func Test_candidate_startElection_withTimeoutsFiring(t *testing.T) {
	raftVoter := fakeVoter{}
	voteStatusChan := make(chan voter.VoteStatus)
	raftVoter.RequestVoteFn = func(term int64) <-chan voter.VoteStatus {
		return voteStatusChan
	}
	f := instantiateDefaultCandidate(candidateStub{fakeVoter: raftVoter})
	ctxt, cancel := context.WithCancel(context.Background())
	ec := f.startElection(ctxt)
	time.AfterFunc(1*time.Second, cancel)
	signal := <-ec
	if signal != struct{}{} {
		t.Error("Expected completion signal but received something else")
	}
}

func instantiateDefaultCandidate(stub candidateStub) candidate {
	//setup the mocks
	if reflect.DeepEqual(stub.fakeTimer, fakeTimer{}) {
		timer := fakeTimer{}
		timer.SetDeadlineFn = func(t time.Time) {
			t.Add(1 * time.Second)
		}
		timer.GetDeadlineFn = func() time.Time {
			return time.Now().Add(-1 * time.Second)
		}
		stub.fakeTimer = timer
	}

	if reflect.DeepEqual(stub.fakeTerm, fakeTerm{}) {
		term := fakeTerm{}
		term.IncrementTermFn = func() int64 {
			return 1
		}
		stub.fakeTerm = term
	}

	if reflect.DeepEqual(stub.fakeVoter, fakeVoter{}) {
		raftVoter := fakeVoter{}
		voteStatusChan := make(chan voter.VoteStatus)
		raftVoter.RequestVoteFn = func(term int64) <-chan voter.VoteStatus {
			go func() {
				defer close(voteStatusChan)
				voteStatusChan <- voter.Leader
			}() //Sending the leader signal
			return voteStatusChan
		}
		stub.fakeVoter = raftVoter
	}
	f := candidate{
		raftTimer:     stub.fakeTimer,
		raftTerm:      stub.fakeTerm,
		raftVoter:     stub.fakeVoter,
		leaderTrigger: stub.leaderTrigger,
	}
	return f
}
