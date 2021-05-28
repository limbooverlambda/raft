package actors

import (
	"context"
	"github.com/kitengo/raft/internal/models"
	raftrpc "github.com/kitengo/raft/internal/rpc"
	"log"
	"sync"
	"testing"
	"time"
)

func Test_leader_Run_ClientReq(t *testing.T) {
	var actualResponse raftrpc.ClientCommandResponse
	expectedResponse := raftrpc.ClientCommandResponse{Committed: true}
	lStub := createLeaderStub()
	lStub.heartbeat.GetSendHeartBeatFn = func() {
		log.Println("Sending heartbeat")
	}
	lStub.raftTimer.GetSetIdleTimeFn = func() {
		log.Println("Setting idle timeout")
	}
	lStub.raftTimer.GetIdleTimoutFn = func() time.Time {
		return time.Now()
	}
	lStub.clientRPC.GetProcessFn = func(meta raftrpc.RaftRpcMeta) (raftrpc.RaftRpcResponse, error) {
		return expectedResponse, nil
	}
	testLeader := leader{
		state:     lStub.state,
		heartbeat: lStub.heartbeat,
		clientRPC: lStub.clientRPC,
		aeRPC:     lStub.aeRPC,
		voteRPC:   lStub.voteRPC,
		raftTerm:  lStub.raftTerm,
		raftTimer: lStub.raftTimer,
	}

	ctxt, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		testLeader.Run(ctxt)
	}()
	respChan := make(chan raftrpc.RaftRpcResponse, 1)
	errChan := make(chan error, 1)
	lStub.clientRequestChan <- raftrpc.ClientCommand{
		Payload:   []byte("payload"),
		RespChan:  respChan,
		ErrorChan: errChan,
	}
	time.AfterFunc(2*time.Second, cancel)
	wg.Wait()
	resp := <-respChan
	actualResponse = resp.(raftrpc.ClientCommandResponse)
	if actualResponse != expectedResponse {
		t.Errorf("expected %v but actual %v", expectedResponse, actualResponse)
	}
}

func Test_leader_Run_AppendEntryReq(t *testing.T) {
	var actualResponse models.AppendEntryResponse
	expectedResponse := models.AppendEntryResponse{
		Term:    1,
		Success: true,
	}
	lStub := createLeaderStub()
	lStub.heartbeat.GetSendHeartBeatFn = func() {
		log.Println("Sending heartbeat")
	}
	lStub.raftTimer.GetSetIdleTimeFn = func() {
		log.Println("Setting idle timeout")
	}
	lStub.raftTimer.GetIdleTimoutFn = func() time.Time {
		return time.Now()
	}
	lStub.raftTerm.GetTermFn = func() int64 {
		return 0
	}
	lStub.aeRPC.GetProcessFn = func(meta raftrpc.RaftRpcMeta) (raftrpc.RaftRpcResponse, error) {
		return models.AppendEntryResponse{
			Term:    1,
			Success: true,
		}, nil
	}

	testLeader := leader{
		state:     lStub.state,
		heartbeat: lStub.heartbeat,
		clientRPC: lStub.clientRPC,
		aeRPC:     lStub.aeRPC,
		voteRPC:   lStub.voteRPC,
		raftTerm:  lStub.raftTerm,
		raftTimer: lStub.raftTimer,
	}

	ctxt, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		testLeader.Run(ctxt)
	}()
	respChan := make(chan raftrpc.RaftRpcResponse, 1)
	errChan := make(chan error, 1)
	lStub.aeRequestChan <- raftrpc.AppendEntry{
		Term:         1,
		LeaderId:     "foo",
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		Entries:      []byte("foo"),
		LeaderCommit: 0,
		RespChan:     respChan,
		ErrorChan:    errChan,
	}
	time.AfterFunc(2*time.Second, cancel)
	wg.Wait()
	resp := <-respChan
	actualResponse = resp.(models.AppendEntryResponse)
	if actualResponse != expectedResponse {
		t.Errorf("expected %v but actual %v", expectedResponse, actualResponse)
	}
}

func Test_leader_Run_ProcessVoteRequest(t *testing.T) {
	var actualResponse raftrpc.RequestVoteResponse
	expectedResponse := raftrpc.RequestVoteResponse{
		Term:        1,
		VoteGranted: true,
	}
	lStub := createLeaderStub()
	lStub.heartbeat.GetSendHeartBeatFn = func() {
		log.Println("Sending heartbeat")
	}
	lStub.raftTimer.GetSetIdleTimeFn = func() {
		log.Println("Setting idle timeout")
	}
	lStub.raftTimer.GetIdleTimoutFn = func() time.Time {
		return time.Now()
	}
	lStub.raftTerm.GetTermFn = func() int64 {
		return 0
	}
	lStub.voteRPC.GetProcessFn = func(meta raftrpc.RaftRpcMeta) (raftrpc.RaftRpcResponse, error) {
		return raftrpc.RequestVoteResponse{
			Term:        1,
			VoteGranted: true,
		}, nil
	}

	testLeader := leader{
		state:     lStub.state,
		heartbeat: lStub.heartbeat,
		clientRPC: lStub.clientRPC,
		aeRPC:     lStub.aeRPC,
		voteRPC:   lStub.voteRPC,
		raftTerm:  lStub.raftTerm,
		raftTimer: lStub.raftTimer,
	}

	ctxt, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		testLeader.Run(ctxt)
	}()
	respChan := make(chan raftrpc.RaftRpcResponse, 1)
	errChan := make(chan error, 1)
	lStub.voteRequestsChan <- raftrpc.RequestVote{
		Term:         0,
		CandidateId:  "",
		LastLogIndex: 0,
		LastLogTerm:  0,
		RespChan:     respChan,
		ErrorChan:    errChan,
	}
	time.AfterFunc(2*time.Second, cancel)
	wg.Wait()
	resp := <-respChan
	actualResponse = resp.(raftrpc.RequestVoteResponse)
	if actualResponse != expectedResponse {
		t.Errorf("expected %v but actual %v", expectedResponse, actualResponse)
	}
}

func Test_leader_Run_HearbeatSender(t *testing.T) {
	var actualHeartbeatCount int
	lStub := createLeaderStub()
	lStub.heartbeat.GetSendHeartBeatFn = func() {
		log.Println("Sending heartbeat")
		actualHeartbeatCount++
	}
	lStub.raftTimer.GetSetIdleTimeFn = func() {
		log.Println("Setting idle timeout")
	}
	lStub.raftTimer.GetIdleTimoutFn = func() time.Time {
		return time.Now().Add(-1 * time.Second)
	}

	testLeader := leader{
		state:     lStub.state,
		heartbeat: lStub.heartbeat,
		clientRPC: lStub.clientRPC,
		aeRPC:     lStub.aeRPC,
		voteRPC:   lStub.voteRPC,
		raftTerm:  lStub.raftTerm,
		raftTimer: lStub.raftTimer,
	}

	ctxt, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		testLeader.Run(ctxt)
	}()

	time.AfterFunc(1*time.Second, cancel)
	wg.Wait()

	if actualHeartbeatCount == 1 {
		t.Errorf("expected more than one send heartbeat but actual %d", actualHeartbeatCount)
	}
}
