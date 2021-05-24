package actors

import (
	"context"
	"github.com/kitengo/raft/internal/models"
	"github.com/kitengo/raft/internal/rpc"
	raftstate "github.com/kitengo/raft/internal/state"
	"log"
	"sync"
	"testing"
	"time"
)

func Test_follower_Run_AppendEntry(t *testing.T) {
	var actualAppendResponse models.AppendEntryResponse
	expectedAppendResponse := models.AppendEntryResponse{
		Term:    1,
		Success: true,
	}
	fStub := createFollowerStub()
	fStub.fakeAeRPC.GetProcessFn = func(meta rpc.RaftRpcMeta) (rpc.RaftRpcResponse, error) {
		ae := meta.(rpc.AppendEntryMeta)
		actualAppendResponse.Term = ae.Term
		actualAppendResponse.Success = true
		return actualAppendResponse, nil
	}
	testFollower := follower{
		state:     fStub.fakeState,
		raftTimer: fStub.fakeTimer,
		aeRPC:     fStub.fakeAeRPC,
		voteRPC:   fStub.fakeVoteRPC,
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		testFollower.Run(ctx)
	}()
	respChan := make(chan rpc.RaftRpcResponse, 1)
	errChan := make(chan error, 1)
	fStub.aeRequestChan <- rpc.AppendEntry{
		Term:     1,
		RespChan: respChan,
		ErrorChan: errChan,
	}
	time.AfterFunc(2 * time.Second, cancel)
	wg.Wait()
	if actualAppendResponse != expectedAppendResponse {
		t.Errorf("expected %v but actual %v", expectedAppendResponse, actualAppendResponse)
	}
}

func Test_follower_Run_VoteRequest(t *testing.T) {
	var actualReqVoteResponse models.RequestVoteResponse
	expectedReqVoteResponse := models.RequestVoteResponse{
		Term:        1,
		VoteGranted: false,
	}
	fStub := createFollowerStub()
	fStub.fakeVoteRPC.GetProcessFn = func(meta rpc.RaftRpcMeta) (rpc.RaftRpcResponse, error) {
		rv := meta.(rpc.RequestVoteMeta)
		actualReqVoteResponse.Term = rv.Term
		actualReqVoteResponse.VoteGranted = false
		return actualReqVoteResponse, nil
	}
	testFollower := follower{
		state:     fStub.fakeState,
		raftTimer: fStub.fakeTimer,
		aeRPC:     fStub.fakeAeRPC,
		voteRPC:   fStub.fakeVoteRPC,
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		testFollower.Run(ctx)
	}()
	respChan := make(chan rpc.RaftRpcResponse, 1)
	errChan := make(chan error, 1)
	fStub.voteRequestsChan <- rpc.RequestVote{
		Term:         1,
		RespChan:     respChan,
		ErrorChan:    errChan,
	}
	time.AfterFunc(2 * time.Second, cancel)
	wg.Wait()
	if actualReqVoteResponse != expectedReqVoteResponse {
		t.Errorf("expected %v but actual %v", expectedReqVoteResponse, actualReqVoteResponse)
	}
}

func Test_follower_Run_DeadlineExceeded(t *testing.T) {
	fStub := createFollowerStub()
	fStub.fakeTimer.GetDeadlineFn = func() time.Time {
		log.Println("Getting deadline")
		return time.Now().Add(-1 * time.Second)
	}
	fStub.fakeTimer.GetSetDeadlineFn = func(currentTime time.Time) {
		log.Println("Setting deadline")
	}
	var actualState raftstate.State
	expectedState := raftstate.CandidateState
	fStub.fakeState.GetSetStateFn = func(state raftstate.State) {
		actualState = state
	}
	testFollower := follower{
		state:     fStub.fakeState,
		raftTimer: fStub.fakeTimer,
		aeRPC:     fStub.fakeAeRPC,
		voteRPC:   fStub.fakeVoteRPC,
	}


	ctx, _ := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		testFollower.Run(ctx)
	}()
	wg.Wait()
	if actualState != expectedState {
		t.Errorf("expected %v but actual %v", expectedState, actualState)
	}
}

