package actors

import (
	"context"
	"github.com/kitengo/raft/internal/member"
	"github.com/kitengo/raft/internal/models"
	raftrpc "github.com/kitengo/raft/internal/rpc"
	raftstate "github.com/kitengo/raft/internal/state"
	raftvoter "github.com/kitengo/raft/internal/voter"
	"sync"
	"testing"
	"time"
)

func Test_candidate_Run_ToLeader(t *testing.T) {
	var actualState raftstate.State
	expectedState := raftstate.LeaderState
	var actualLeaderID string
	expectedLeaderID := "foo"

	cStub := createCandidateStub()
	cStub.fakeMember.GetSetLeaderIDFn = func(leaderID string) {
		actualLeaderID = leaderID
	}
	cStub.fakeMember.GetSelfFn = func() member.Entry {
		return member.Entry{
			ID: expectedLeaderID,
		}
	}
	cStub.fakeState.GetSetStateFn = func(state raftstate.State) {
		actualState = state
	}
	voteStatuses := cStub.voteStatusChan

	testCandidate := candidate{
		member:    cStub.fakeMember,
		state:     cStub.fakeState,
		term:      cStub.fakeTerm,
		voter:     cStub.fakeVoter,
		raftTimer: cStub.fakeTimer,
		aeRPC:     cStub.fakeAeRPC,
		voteRPC:   cStub.fakeVoteRPC,
	}
	cctx, _ := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		testCandidate.Run(cctx)
	}()
	voteStatuses <- raftvoter.Leader
	wg.Wait()
	if actualState != expectedState {
		t.Errorf("expected %v but actual %v", expectedState, actualState)
	}
	if actualLeaderID != expectedLeaderID {
		t.Errorf("expected %v but actual %v", expectedLeaderID, actualLeaderID)
	}
}

func Test_candidate_Run_ToFollower(t *testing.T) {
	var actualState raftstate.State
	expectedState := raftstate.FollowerState

	cStub := createCandidateStub()

	cStub.fakeState.GetSetStateFn = func(state raftstate.State) {
		actualState = state
	}
	voteStatuses := cStub.voteStatusChan

	testCandidate := candidate{
		member:    cStub.fakeMember,
		state:     cStub.fakeState,
		term:      cStub.fakeTerm,
		voter:     cStub.fakeVoter,
		raftTimer: cStub.fakeTimer,
		aeRPC:     cStub.fakeAeRPC,
		voteRPC:   cStub.fakeVoteRPC,
	}
	cctx, _ := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		testCandidate.Run(cctx)
	}()
	voteStatuses <- raftvoter.Follower
	wg.Wait()
	if actualState != expectedState {
		t.Errorf("expected %v but actual %v", expectedState, actualState)
	}
}

func Test_candidate_Run_ToCandidate(t *testing.T) {
	var actualState raftstate.State
	expectedState := raftstate.CandidateState

	cStub := createCandidateStub()

	cStub.fakeState.GetSetStateFn = func(state raftstate.State) {
		actualState = state
	}
	voteStatuses := cStub.voteStatusChan

	testCandidate := candidate{
		member:    cStub.fakeMember,
		state:     cStub.fakeState,
		term:      cStub.fakeTerm,
		voter:     cStub.fakeVoter,
		raftTimer: cStub.fakeTimer,
		aeRPC:     cStub.fakeAeRPC,
		voteRPC:   cStub.fakeVoteRPC,
	}

	cctx, cancel := context.WithCancel(context.Background())
	go func() {
		testCandidate.Run(cctx)
	}()
	voteStatuses <- raftvoter.Split
	<-time.After(1 * time.Second)
	cancel()
	if actualState != expectedState {
		t.Errorf("expected %v but actual %v", expectedState, actualState)
	}
}

func Test_candidate_Run_AE_WithHigherTerm(t *testing.T) {
	var actualState raftstate.State
	expectedState := raftstate.FollowerState

	var actualLeaderID string
	expectedLeaderID := "foo"
	cStub := createCandidateStub()
	cStub.fakeAeRPC.GetProcessFn = func(meta raftrpc.RaftRpcMeta) (raftrpc.RaftRpcResponse,
		error) {
		return models.AppendEntryResponse{
			Term:    2,
			Success: true,
		}, nil
	}

	cStub.fakeMember.GetSetLeaderIDFn = func(leaderID string) {
		actualLeaderID = leaderID
	}

	aeChan := cStub.aeRequestChan

	testCandidate := candidate{
		member:    cStub.fakeMember,
		state:     cStub.fakeState,
		term:      cStub.fakeTerm,
		voter:     cStub.fakeVoter,
		raftTimer: cStub.fakeTimer,
		aeRPC:     cStub.fakeAeRPC,
		voteRPC:   cStub.fakeVoteRPC,
	}

	cctx, _ := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		testCandidate.Run(cctx)
	}()
	respChan := make(chan raftrpc.RaftRpcResponse, 1)
	errChan := make(chan error, 1)
	aeChan <- raftrpc.AppendEntry{
		Term:      2,
		LeaderId:  expectedLeaderID,
		RespChan:  respChan,
		ErrorChan: errChan,
	}
	wg.Wait()
	if actualState != expectedState {
		t.Errorf("expected %v but actual %v", expectedState, actualState)
	}
	if actualLeaderID != expectedLeaderID {
		t.Errorf("expected %v but actual %v", expectedLeaderID, actualLeaderID)
	}
}

func Test_candidate_Run_VoteRequest(t *testing.T) {
	var actualState raftstate.State
	expectedState := raftstate.CandidateState

	cStub := createCandidateStub()
	cStub.fakeVoteRPC.GetProcessFn = func(meta raftrpc.RaftRpcMeta) (raftrpc.RaftRpcResponse, error) {
		return models.RequestVoteResponse{
			Term:        2,
			VoteGranted: true,
		}, nil
	}
	cStub.fakeState.GetSetStateFn = func(state raftstate.State) {
		actualState = state
	}
	voteRequests := cStub.voteRequestsChan
	testCandidate := candidate{
		member:    cStub.fakeMember,
		state:     cStub.fakeState,
		term:      cStub.fakeTerm,
		voter:     cStub.fakeVoter,
		raftTimer: cStub.fakeTimer,
		aeRPC:     cStub.fakeAeRPC,
		voteRPC:   cStub.fakeVoteRPC,
	}

	cctx, cancel := context.WithCancel(context.Background())
	//var wg sync.WaitGroup
	//wg.Add(1)
	go func() {
		//defer wg.Done()
		testCandidate.Run(cctx)
	}()
	respChan := make(chan raftrpc.RaftRpcResponse, 1)
	errChan := make(chan error, 1)
	voteRequests <- raftrpc.RequestVote{
		Term:         0,
		CandidateId:  "",
		LastLogIndex: 0,
		LastLogTerm:  0,
		RespChan:     respChan,
		ErrorChan:    errChan,
	}
	<-time.After(1 * time.Second)
	cancel()
	if actualState != expectedState {
		t.Errorf("expected %v but actual %v", expectedState, actualState)
	}
}


