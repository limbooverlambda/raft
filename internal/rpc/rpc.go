package rpc

import (
	"log"
)

type AppendEntryRequest struct{}
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
	RespChan     chan<- AppendEntryResponse
	ErrorChan    chan<- error
}

type AppendEntryMeta struct {
	Term         int64
	LeaderId     string
	PrevLogIndex int64
	PrevLogTerm  int64
	Entries      []byte
	LeaderCommit int64
}

type RaftAppendEntry interface {
	ReceiveAppendEntry(appendEntry AppendEntry)
	Process(meta AppendEntryMeta) (AppendEntryResponse, error)
	AppendEntryReqChan() <-chan AppendEntry
}

func NewRaftAppendEntry() RaftAppendEntry {
	appendEntryChan := make(chan AppendEntry, 1)
	return raftAppendEntry{
		appendEntryChan: appendEntryChan,
	}
}

type raftAppendEntry struct{
	appendEntryChan chan AppendEntry
}

func (ra raftAppendEntry) ReceiveAppendEntry(appendEntry AppendEntry) {
	ra.appendEntryChan <- appendEntry
}

func (ra raftAppendEntry) Process(meta AppendEntryMeta) (AppendEntryResponse, error) {
	panic("implement me")
}

func (ra raftAppendEntry) AppendEntryReqChan() <-chan AppendEntry {
	return ra.appendEntryChan
}

type RequestVote struct {
	Term         int64
	CandidateId  string
	LastLogIndex int64
	LastLogTerm  int64
	RespChan     chan<- RequestVoteResponse
	ErrorChan    chan<- error
}

type RequestVoteMeta struct {
	Term         int64
	CandidateId  string
	LastLogIndex int64
	LastLogTerm  int64
}

type RequestVoteResponse struct{ Term int64 }

type RaftRequestVote interface {
	ReceiveRequestVote(requestVote RequestVote)
	Process(meta RequestVoteMeta) (RequestVoteResponse, error)
	RequestVoteReqChan() <-chan RequestVote
}

func NewRaftRequestVote() RaftRequestVote {
	requestVoteChan := make(chan RequestVote, 1)
	return raftRequestVote{
		requestVoteChan: requestVoteChan,
	}
}

type raftRequestVote struct{
	requestVoteChan chan RequestVote
}

func (rrV raftRequestVote) ReceiveRequestVote(requestVote RequestVote) {
	rrV.requestVoteChan <- requestVote
}

func (raftRequestVote) Process(meta RequestVoteMeta) (RequestVoteResponse, error) {
	panic("implement me")
}

func (rrV raftRequestVote) RequestVoteReqChan() <-chan RequestVote {
	return rrV.requestVoteChan
}

type ClientCommand struct {
	Payload   []byte
	RespChan  chan<- ClientCommandResponse
	ErrorChan chan<- error
}

type ClientCommandMeta struct {
	Payload []byte
}

type ClientCommandResponse struct {
	Committed bool
}

type RaftClientCommand interface {
	ReceiveClientCommand(clientCommand ClientCommand)
	Process(meta ClientCommandMeta) (ClientCommandResponse, error)
	ClientCommandReqChan() <-chan ClientCommand
}

func NewRaftClientCommand() RaftClientCommand {
	clientCommandChan := make(chan ClientCommand, 1)
	return &raftClientCommand{
		clientCommandChan: clientCommandChan,
	}
}

type raftClientCommand struct{
	clientCommandChan chan ClientCommand
}

func (rcc *raftClientCommand) ReceiveClientCommand(clientCommand ClientCommand) {
	log.Printf("Received clientCommand %v\n", clientCommand)
	rcc.clientCommandChan <- clientCommand
}

func (*raftClientCommand) Process(meta ClientCommandMeta) (ClientCommandResponse, error) {
	log.Printf("Processing client request %v\n", meta)
	return ClientCommandResponse{Committed: true}, nil
}

func (rcc *raftClientCommand) ClientCommandReqChan() <-chan ClientCommand {
	return rcc.clientCommandChan
}
