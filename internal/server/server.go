package server

import (
	"bytes"
	"encoding/gob"

	"github.com/kitengo/raft/internal/locator"
	raftmodels "github.com/kitengo/raft/internal/models"
	raftrpc "github.com/kitengo/raft/internal/rpc"
)

type ResponseChan <-chan raftmodels.Response
type ErrorChan <-chan error

//The RaftServer will
type RaftServer interface {
	Accept(request raftmodels.Request) (ResponseChan, ErrorChan)
}

func NewRaftServer(locator locator.RpcLocator) RaftServer {
	return &raftServer{
		appendEntrySvc:   locator.GetAppendEntrySvc(),
		requestVoteSvc:   locator.GetRequestVoteSvc(),
		clientCommandSvc: locator.GetClientCommandSvc(),
	}
}

type raftServer struct {
	appendEntrySvc   raftrpc.RaftRpc
	requestVoteSvc   raftrpc.RaftRpc
	clientCommandSvc raftrpc.RaftRpc
}

//Accept responds to a request made to a raft server
//The first thing it does is to demux to understand
//what kind of incoming request it is.
//The request can be one of:
// - AppendEntry (actual appends and heartbeats)
// - RequestVote
// - ClientCommand
// Once the request type has been retrieved, The request
// is wrapped in an appropriate struct for the action
//  An example would be something along the lines of:
//   The incoming request is AppendEntry
//     We create the following struct for the AppendEntry request
//     raftrpc.AppendEntry {
//          term int64
//          leaderId string
//          prevLogIndex int64
//          prevLogTerm int64
//          entries []byte
//			leaderCommit int64
//          RespChan chan <- AppendEntryResp
//          ErrorChan chan <- error
//     }
// After the struct has been created, the appropriate Actor function will be
// called for one of the state machines to work on this payload. In this case
// raftrpc.ProcessAppendEntry(ae AppendEntry) which will enqueue the AppendEntry
// to a channel maintained by the rpc package.
// The appropriate worker (leader, follower, candidate) will pick up the
// appendEntry and process it appropriately.
// Once it is done processing it, it will send either the response or the
// error back on the RespChan or the Error chan
// The same pattern will be leveraged by RequestVote
func (rs *raftServer) Accept(request raftmodels.Request) (respChan ResponseChan, errChan ErrorChan) {
	switch request.RequestType {
	case raftmodels.AppendEntry:
		{
			respChan, errChan = rs.appendEntry(request.Payload)
		}
	case raftmodels.RequestVote:
		{
			respChan, errChan = rs.requestVote(request.Payload)
		}
	case raftmodels.ClientCommand:
		{
			respChan, errChan = rs.clientCommand(request.Payload)
		}
	}
	return
}

func (rs *raftServer) appendEntry(payload []byte) (ResponseChan, ErrorChan) {
	extractPayloadFunc := func(payload []byte) (interface{}, error) {
		decoder := gob.NewDecoder(bytes.NewBuffer(payload))
		var aePayload raftmodels.AppendEntryPayload
		err := decoder.Decode(&aePayload)
		return aePayload, err
	}
	toRaftRpcRequest := func(
		payload interface{},
		aeRespChan chan raftrpc.RaftRpcResponse,
		errChan chan error,
	) raftrpc.RaftRpcRequest {
		aePayload := payload.(raftmodels.AppendEntryPayload)
		return raftrpc.AppendEntry{
			Term:         aePayload.Term,
			LeaderId:     aePayload.LeaderId,
			PrevLogIndex: aePayload.PrevLogIndex,
			PrevLogTerm:  aePayload.PrevLogTerm,
			Entries:      aePayload.Entries,
			LeaderCommit: aePayload.LeaderCommit,
			RespChan:     aeRespChan,
			ErrorChan:    errChan,
		}
	}
	receiveFn := func(ae raftrpc.RaftRpcRequest) {
		rs.appendEntrySvc.Receive(ae)
	}

	return rs.processRequest(payload, extractPayloadFunc, toRaftRpcRequest, receiveFn)
}

func (rs *raftServer) requestVote(payload []byte) (ResponseChan, ErrorChan) {
	extractPayloadFn := func(payload []byte) (interface{}, error) {
		decoder := gob.NewDecoder(bytes.NewBuffer(payload))
		var rvPayload raftmodels.RequestVotePayload
		err := decoder.Decode(&rvPayload)
		return rvPayload, err
	}

	toRaftRpcRequestFn := func(
		payload interface{},
		respChan chan raftrpc.RaftRpcResponse,
		errChan chan error,
	) raftrpc.RaftRpcRequest {
		rvPayload := payload.(raftmodels.RequestVotePayload)
		return raftrpc.RequestVote{
			Term:         rvPayload.Term,
			CandidateId:  rvPayload.CandidateId,
			LastLogIndex: rvPayload.LastLogIndex,
			LastLogTerm:  rvPayload.LastLogTerm,
			RespChan:     respChan,
			ErrorChan:    errChan,
		}
	}

	receiveFn := func(req raftrpc.RaftRpcRequest) {
		rs.requestVoteSvc.Receive(req)
	}

	return rs.processRequest(
		payload,
		extractPayloadFn,
		toRaftRpcRequestFn,
		receiveFn,
	)
}

func (rs *raftServer) clientCommand(payload []byte) (ResponseChan, ErrorChan) {
	extractPayloadFn := func(payload []byte) (interface{}, error) {
		decoder := gob.NewDecoder(bytes.NewBuffer(payload))
		var clPayload raftmodels.ClientCommandPayload
		err := decoder.Decode(&clPayload)
		return clPayload, err
	}

	toRaftRpcRequestFn := func(
		payload interface{},
		respChan chan raftrpc.RaftRpcResponse,
		errChan chan error,
	) raftrpc.RaftRpcRequest {
		clPayload := payload.(raftmodels.ClientCommandPayload)
		return raftrpc.ClientCommand{
			Payload:   clPayload.ClientCommand,
			RespChan:  respChan,
			ErrorChan: errChan,
		}
	}

	receiveFn := func(req raftrpc.RaftRpcRequest) {
		rs.clientCommandSvc.Receive(req)
	}

	return rs.processRequest(
		payload,
		extractPayloadFn,
		toRaftRpcRequestFn,
		receiveFn,
	)
}

type extractPayloadFn func(payload []byte) (interface{}, error)

type toRaftRpcRequestFn func(
	payload interface{},
	aeRespChan chan raftrpc.RaftRpcResponse,
	errChan chan error,
) raftrpc.RaftRpcRequest

type receiveFn func(raftrpc.RaftRpcRequest)

func (rs *raftServer) processRequest(payload []byte,
	extractPayloadFunc extractPayloadFn,
	toRaftRpcRequest toRaftRpcRequestFn,
	receiveFn receiveFn) (ResponseChan, ErrorChan) {
	respChan := make(chan raftmodels.Response)
	errChan := make(chan error)

	go func() {
		defer close(respChan)
		defer close(errChan)
		payload, err := extractPayloadFunc(payload)
		if err != nil {
			errChan <- err
			return
		}
		rRespChan := make(chan raftrpc.RaftRpcResponse)
		req := toRaftRpcRequest(payload, rRespChan, errChan)
		receiveFn(req)
		select {
		case resp := <-rRespChan:
			{
				var payloadBytes bytes.Buffer
				enc := gob.NewEncoder(&payloadBytes)
				err := enc.Encode(resp)
				if err != nil {
					errChan <- err
					return
				}
				respChan <- raftmodels.Response{Payload: payloadBytes.Bytes()}
				return
			}
		}
	}()
	return respChan, errChan
}
