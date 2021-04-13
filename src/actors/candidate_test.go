package actors

import (
	"context"
	"reflect"
	"testing"
	"time"

	"kitengo/raft/rpc"
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

func Test_candidate_startCandidateWithCancellation(t *testing.T) {
	appendEntry := fakeAppendEntry{}
	aeChan := make(chan rpc.AppendEntryRequest)
	appendEntry.AppendEntryReqChanFn = func() <-chan rpc.AppendEntryRequest {
		return aeChan
	}
	f := instantiateDefaultCandidate(candidateStub{
		fakeAppendEntry: appendEntry,
	})
	ctxt, cancel := context.WithCancel(context.Background())
	f.context = ctxt
	time.AfterFunc(5*time.Second, cancel)
	c := f.startCandidate()
	<-c
}

func Test_candidate_startCandidateWithAppendEntrySuccess(t *testing.T) {
	appendEntry := fakeAppendEntry{}
	aeChan := make(chan rpc.AppendEntryRequest)
	followerTrigger := make(chan FollowerTrigger, 1)
	appendEntry.AppendEntryReqChanFn = func() <-chan rpc.AppendEntryRequest {
		go func() {
			aeChan <- rpc.AppendEntryRequest{}
		}()
		return aeChan
	}
	appendEntry.ProcessFn = func(request rpc.AppendEntryRequest) rpc.AppendEntryResponse {
		return rpc.AppendEntryResponse{
			Success: true,
		}
	}
	respChan := make(chan rpc.AppendEntryResponse, 1)
	appendEntry.AppendEntryRespChanFn = func() chan<- rpc.AppendEntryResponse {
		return respChan
	}

	rvoter := fakeVoter{}
	rvoter.RequestVoteFn = func(term int64) <-chan voter.VoteStatus {
		vChan := make(chan voter.VoteStatus)
		defer close(vChan)
		time.Sleep(10 * time.Second)
		return vChan
	}
	f := instantiateDefaultCandidate(candidateStub{
		fakeAppendEntry: appendEntry,
		followerTrigger: followerTrigger,
		fakeVoter:       rvoter,
	})
	ctxt, cancel := context.WithCancel(context.Background())
	f.context = ctxt
	time.AfterFunc(5*time.Second, cancel)
	c := f.startCandidate()
	<-c
	<-followerTrigger

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
	<-ec
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
		followerTrigger: stub.followerTrigger,
		leaderTrigger:   stub.leaderTrigger,
		raftTimer:       stub.fakeTimer,
		raftTerm:        stub.fakeTerm,
		raftVoter:       stub.fakeVoter,
		raftAppendEntry: stub.fakeAppendEntry,
	}
	return f
}
