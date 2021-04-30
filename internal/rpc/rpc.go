package rpc

import (
	"errors"
	"github.com/kitengo/raft/internal/appendentry"
	"github.com/kitengo/raft/internal/applicator"
	raftlog "github.com/kitengo/raft/internal/log"
	"github.com/kitengo/raft/internal/member"
	"github.com/kitengo/raft/internal/models"
	raftstate "github.com/kitengo/raft/internal/state"
	"github.com/kitengo/raft/internal/term"
	"log"
	"time"
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



type AppendEntry struct {
	Term         int64
	LeaderId     string
	PrevLogIndex uint64
	PrevLogTerm  int64
	Entries      []byte
	LeaderCommit uint64
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
	PrevLogIndex uint64
	PrevLogTerm  int64
	Entries      []byte
	LeaderCommit uint64
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
	return models.AppendEntryResponse{
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

func NewRaftClientCommand(
	raftTerm term.RaftTerm,
	raftLog raftlog.RaftLog,
	raftMember member.RaftMember,
	sender appendentry.Sender,
	index *raftstate.RaftIndex,
	raftApplicator applicator.RaftApplicator) RaftRpc {
	clientCommandChan := make(chan RaftRpcRequest, 1)
	return &raftClientCommand{
		clientCommandChan: clientCommandChan,
		raftTerm:          raftTerm,
		raftLog:           raftLog,
		raftMember:        raftMember,
		appendEntrySender: sender,
		raftIndex:         index,
		raftApplicator:    raftApplicator,
	}
}

type raftClientCommand struct {
	clientCommandChan chan RaftRpcRequest
	raftTerm          term.RaftTerm
	raftLog           raftlog.RaftLog
	raftMember        member.RaftMember
	appendEntrySender appendentry.Sender
	raftIndex         *raftstate.RaftIndex
	raftApplicator    applicator.RaftApplicator
}

func (rcc *raftClientCommand) Receive(request RaftRpcRequest) {
	log.Printf("Received clientCommand %v\n", request)
	clientCommand := request.(ClientCommand)
	rcc.clientCommandChan <- clientCommand
}

//Process The leader
//appends the command to its log as a new entry, then issues AppendEntries RPCs in
//parallel to each of the other servers to replicate the entry. When the entry has been
//safely replicated to the majority of the servers, the leader applies
//the entry to its state machine and returns the result of that
//execution to the client. If followers crash or run slowly,
//or if network packets are lost, the leader retries AppendEntries RPCs indefinitely
func (rcc *raftClientCommand) Process(meta RaftRpcMeta) (RaftRpcResponse, error) {
	log.Printf("Processing client request %v\n", meta)
	//Append the entry to the log
	clientCommandMeta := meta.(ClientCommandMeta)
	term := rcc.raftTerm.GetTerm()
	leader := rcc.raftMember.Leader()
	payload := clientCommandMeta.Payload
	aer, err := rcc.raftLog.AppendEntry(raftlog.Entry{
		Term:    term,
		Payload: payload,
	})
	if err != nil {
		log.Printf("unable to append entry to log due to %v\n", err)
		return nil, err
	}

	members, err := rcc.raftMember.List()
	if err != nil {
		log.Printf("unable to list members of the cluster %v\n", err)
		return nil, err
	}
	//Send the appendEntry requests to all the peers
	respChan := make(chan models.AppendEntryResponse, len(members))
	for _, member := range members {
		rcc.appendEntrySender.ForwardEntry(appendentry.Entry{
			MemberID:     member.ID,
			RespChan:     respChan,
			Term:         term,
			LeaderID:     leader.ID,
			PrevLogIndex: aer.PrevLogOffset,
			PrevLogTerm:  aer.PrevLogTerm,
			Entries:      payload,
			LeaderCommit: aer.LogOffset,
			MemberAddr:   member.Address,
		})
	}
	if err = rcc.waitForMajorityAcks(len(members), 100*time.Millisecond, respChan); err != nil {
		return ClientCommandResponse{Committed: false}, nil
	}
	//If majority votes are received, set the commit index
	rcc.raftIndex.SetCommitOffset(aer.LogOffset)
	//Apply the command
	if err = rcc.raftApplicator.Apply(payload); err == nil {
		//Upon successful application, set the applyIndex
		rcc.raftIndex.SetApplyOffset(aer.LogOffset)
	}
	//send the ClientCommandResponse with Committed set to true
	return ClientCommandResponse{Committed: true}, nil
}

func (rcc *raftClientCommand) waitForMajorityAcks(memberCount int,
	deadlineTime time.Duration,
	respChan chan models.AppendEntryResponse) (err error) {
	deadlineChan := time.After(deadlineTime)
	var voteCount int
	majorityCount := (memberCount >> 1) + 1
	for {
		select {
		case t := <-deadlineChan:
			{
				log.Println("deadline exceeded", t)
				err = errors.New("deadline exceeded")
				return
			}
		case resp := <-respChan:
			{
				if resp.Success {
					voteCount++
				}
				if voteCount >= majorityCount {
					return
				}
			}
		}
	}
}

func (rcc *raftClientCommand) RaftRpcReqChan() <-chan RaftRpcRequest {
	return rcc.clientCommandChan
}
