package server

import (
	"bytes"
	"encoding/gob"

	"kitengo/raft/rpc"
)

//RequestType defines the kind of request received
//by the RaftServer, it can be one of 'AppendEntry'
// or 'RequestVote'
type RequestType int

const (
	AppendEntry RequestType = iota
	RequestVote
)

type Request struct {
	RequestType RequestType
	Payload     []byte
}

type AppendEntryPayload struct {
	Term         int64
	LeaderId     string
	PrevLogIndex int64
	PrevLogTerm  int64
	Entries      []byte
	LeaderCommit int64
}

type RequestVotePayload struct {
	Term         int64
	CandidateId  string
	LastLogIndex int64
	LastLogTerm  int64
}

type Response struct {
	Payload []byte
}

type ResponseChan <-chan Response
type ErrorChan <-chan error

//The RaftServer will
type RaftServer interface {
	Accept(request Request) (ResponseChan, ErrorChan)
}

type raftServer struct {
	requestChan    chan<- Request
	appendEntrySvc rpc.RaftAppendEntry
	requestVoteSvc rpc.RaftRequestVote
}

//Accept responds to a request made to a raft server
//The first thing it does is to demux to understand
//what kind of incoming request it is.
//The request can be one of:
// - AppendEntry (actual appends and heartbeats)
// - RequestVote
// Once the request type has been retrieved, The request
// is wrapped in an appropriate struct for the action
//  An example would be something along the lines of:
//   The incoming request is AppendEntry
//     We create the following struct for the AppendEntry request
//     rpc.AppendEntry {
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
// rpc.ProcessAppendEntry(ae AppendEntry) which will enqueue the AppendEntry
// to a channel maintained by the rpc package.
// The appropriate worker (leader, follower, candidate) will pick up the
// appendEntry and process it appropriately.
// Once it is done processing it, it will send either the response or the
// error back on the RespChan or the Error chan
// The same pattern will be leveraged by RequestVote
func (rs *raftServer) Accept(request Request) (respChan ResponseChan, errChan ErrorChan) {
	switch request.RequestType {
	case AppendEntry:
		{
			respChan, errChan = rs.appendEntry(request.Payload)
		}
	case RequestVote:
		{
			respChan, errChan = rs.requestVote(request.Payload)
		}
	}
	return
}

func (rs *raftServer) appendEntry(payload []byte) (ResponseChan, ErrorChan) {
	respChan := make(chan Response)
	errChan := make(chan error)
	go func() {
		defer close(respChan)
		defer close(errChan)
		//TODO: Not the right spot for decoding.
		decoder := gob.NewDecoder(bytes.NewBuffer(payload))
		var aePayload AppendEntryPayload
		err := decoder.Decode(aePayload)
		if err != nil {
			errChan <- err
			return
		}
		aeRespChan := make(chan rpc.AppendEntryResponse)
		ae := rpc.AppendEntry{
			Term:         aePayload.Term,
			LeaderId:     aePayload.LeaderId,
			PrevLogIndex: aePayload.PrevLogIndex,
			PrevLogTerm:  aePayload.PrevLogTerm,
			Entries:      aePayload.Entries,
			LeaderCommit: aePayload.LeaderCommit,
			RespChan:     aeRespChan,
			ErrorChan:    errChan,
		}
		rs.appendEntrySvc.ReceiveAppendEntry(ae)
		select {
		case resp := <-aeRespChan:
			{
				var payloadBytes bytes.Buffer
				enc := gob.NewEncoder(&payloadBytes)
				err := enc.Encode(resp)
				if err != nil {
					errChan <- err
					return
				}
				respChan <- Response{Payload: payloadBytes.Bytes()}
				close(respChan)
				return
			}
		}
	}()
	return respChan, errChan
}

func (rs *raftServer) requestVote(payload []byte) (ResponseChan, ErrorChan) {
	respChan := make(chan Response)
	errChan := make(chan error)
	go func() {
		defer close(respChan)
		defer close(errChan)
		//TODO: Not the right spot for decoding.
		decoder := gob.NewDecoder(bytes.NewBuffer(payload))
		var rvPayload RequestVotePayload
		err := decoder.Decode(rvPayload)
		if err != nil {
			errChan <- err
			return
		}
		rvRespChan := make(chan rpc.RequestVoteResponse)
		rv := rpc.RequestVote{
			RespChan:  rvRespChan,
			ErrorChan: errChan,
		}
		rs.requestVoteSvc.ReceiveRequestVote(rv)
		select {
		case resp := <-rvRespChan:
			{
				var payloadBytes bytes.Buffer
				enc := gob.NewEncoder(&payloadBytes)
				err := enc.Encode(resp)
				if err != nil {
					errChan <- err
					return
				}
				respChan <- Response{Payload: payloadBytes.Bytes()}
				close(respChan)
				return
			}
		}
	}()
	return respChan, errChan
}
