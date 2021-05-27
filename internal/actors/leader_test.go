package actors

import (
	"context"
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
