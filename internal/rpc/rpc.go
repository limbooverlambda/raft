package rpc

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
