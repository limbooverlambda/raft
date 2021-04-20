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
	return raftAppendEntry{}
}

type raftAppendEntry struct{}

func (raftAppendEntry) ReceiveAppendEntry(appendEntry AppendEntry) {
	panic("implement me")
}

func (raftAppendEntry) Process(meta AppendEntryMeta) (AppendEntryResponse, error) {
	panic("implement me")
}

func (raftAppendEntry) AppendEntryReqChan() <-chan AppendEntry {
	panic("implement me")
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
	return raftRequestVote{}
}

type raftRequestVote struct{}

func (raftRequestVote) ReceiveRequestVote(requestVote RequestVote) {
	panic("implement me")
}

func (raftRequestVote) Process(meta RequestVoteMeta) (RequestVoteResponse, error) {
	panic("implement me")
}

func (raftRequestVote) RequestVoteReqChan() <-chan RequestVote {
	panic("implement me")
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
	return raftClientCommand{}
}

type raftClientCommand struct{}

func (raftClientCommand) ReceiveClientCommand(clientCommand ClientCommand) {
	log.Printf("Received clientCommand %v\n", clientCommand)
}

func (raftClientCommand) Process(meta ClientCommandMeta) (ClientCommandResponse, error) {
	panic("implement me")
}

func (raftClientCommand) ClientCommandReqChan() <-chan ClientCommand {
	panic("implement me")
}
