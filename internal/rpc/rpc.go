package rpc

import (
	"log"
)

type RaftRpc interface {
	Receive(request RaftRpcRequest)
	Process(meta RaftRpcMeta) (RaftRpcResponse, error)
	RaftRpcReqChan() <-chan RaftRpcRequest
}

type RaftRpcRequest interface {
	GetResponseChan() chan<- RaftRpcResponse
	GetErrorChan() chan<- error
}
type RaftRpcMeta interface{}
type RaftRpcResponse interface{}

type AppendEntryResponse struct {
	Term    int64
	Success bool
}

type AppendEntry struct {
	Term         int64
	LeaderId     string
	PrevLogIndex int64
	PrevLogTerm  int64
	Entries      []byte
	LeaderCommit int64
	RespChan     chan<- RaftRpcResponse
	ErrorChan    chan<- error
}

func (a AppendEntry) GetResponseChan() chan<- RaftRpcResponse {
	return a.RespChan
}

func (a AppendEntry) GetErrorChan() chan<- error {
	return a.ErrorChan
}

type AppendEntryMeta struct {
	Term         int64
	LeaderId     string
	PrevLogIndex int64
	PrevLogTerm  int64
	Entries      []byte
	LeaderCommit int64
}

func NewRaftAppendEntry() RaftRpc {
	appendEntryChan := make(chan RaftRpcRequest, 1)
	return raftAppendEntry{
		appendEntryChan: appendEntryChan,
	}
}

type raftAppendEntry struct {
	appendEntryChan chan RaftRpcRequest
}

func (ra raftAppendEntry) Receive(request RaftRpcRequest) {
	appendEntry := request.(AppendEntry)
	ra.appendEntryChan <- appendEntry
}

func (ra raftAppendEntry) Process(meta RaftRpcMeta) (RaftRpcResponse, error) {
	log.Printf("Processing AppendEntry %v\n", meta)
	return AppendEntryResponse{
		Term:    1,
		Success: false,
	}, nil
}

func (ra raftAppendEntry) RaftRpcReqChan() <-chan RaftRpcRequest {
	return ra.appendEntryChan
}

type RequestVote struct {
	Term         int64
	CandidateId  string
	LastLogIndex int64
	LastLogTerm  int64
	RespChan     chan<- RaftRpcResponse
	ErrorChan    chan<- error
}

func (a RequestVote) GetResponseChan() chan<- RaftRpcResponse {
	return a.RespChan
}

func (a RequestVote) GetErrorChan() chan<- error {
	return a.ErrorChan
}

type RequestVoteMeta struct {
	Term         int64
	CandidateId  string
	LastLogIndex int64
	LastLogTerm  int64
}

type RequestVoteResponse struct{ Term int64 }

func NewRaftRequestVote() RaftRpc {
	requestVoteChan := make(chan RaftRpcRequest, 1)
	return raftRequestVote{
		requestVoteChan: requestVoteChan,
	}
}

type raftRequestVote struct {
	requestVoteChan chan RaftRpcRequest
}

func (rrV raftRequestVote) Receive(request RaftRpcRequest) {
	requestVote := request.(RequestVote)
	rrV.requestVoteChan <- requestVote
}

func (rrV raftRequestVote) Process(meta RaftRpcMeta) (RaftRpcResponse, error) {
	log.Println("Received request vote", meta)
	return RequestVoteResponse{Term: 1}, nil
}

func (rrV raftRequestVote) RaftRpcReqChan() <-chan RaftRpcRequest {
	return rrV.requestVoteChan
}

func (rrV raftRequestVote) ReceiveRequestVote(requestVote RequestVote) {
	rrV.requestVoteChan <- requestVote
}

type ClientCommand struct {
	Payload   []byte
	RespChan  chan<- RaftRpcResponse
	ErrorChan chan<- error
}

func (a ClientCommand) GetResponseChan() chan<- RaftRpcResponse {
	return a.RespChan
}

func (a ClientCommand) GetErrorChan() chan<- error {
	return a.ErrorChan
}

type ClientCommandMeta struct {
	Payload []byte
}

type ClientCommandResponse struct {
	Committed bool
}

func NewRaftClientCommand() RaftRpc {
	clientCommandChan := make(chan RaftRpcRequest, 1)
	return &raftClientCommand{
		clientCommandChan: clientCommandChan,
	}
}

type raftClientCommand struct {
	clientCommandChan chan RaftRpcRequest
}

func (rcc *raftClientCommand) Receive(request RaftRpcRequest) {
	log.Printf("Received clientCommand %v\n", request)
	clientCommand := request.(ClientCommand)
	rcc.clientCommandChan <- clientCommand
}

func (rcc *raftClientCommand) Process(meta RaftRpcMeta) (RaftRpcResponse, error) {
	log.Printf("Processing client request %v\n", meta)
	return ClientCommandResponse{Committed: true}, nil
}

func (rcc *raftClientCommand) RaftRpcReqChan() <-chan RaftRpcRequest {
	return rcc.clientCommandChan
}
