package rpc

type AppendEntryRequest struct{}
type AppendEntryResponse struct {
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

type RaftAppendEntry interface {
	ReceiveAppendEntry(appendEntry AppendEntry)
	AppendEntryReqChan() <-chan AppendEntryRequest
	Process(request AppendEntryRequest) AppendEntryResponse
	AppendEntryRespChan() chan<- AppendEntryResponse
}

type RequestVote struct {
	Term         int64
	CandidateId  string
	LastLogIndex int64
	LastLogTerm  int64
	RespChan     chan<- RequestVoteResponse
	ErrorChan    chan<- error
}

type RequestVoteResponse struct{}

type RaftRequestVote interface {
	ReceiveRequestVote(requestVote RequestVote)
}
