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
	return &raftAppendEntry{
		appendEntryChan: appendEntryChan,
	}
}

type raftAppendEntry struct {
	raftTerm        term.RaftTerm
	raftLog         raftlog.RaftLog
	raftIndex       raftstate.RaftIndex
	appendEntryChan chan RaftRpcRequest
}

func (ra *raftAppendEntry) Receive(request RaftRpcRequest) {
	appendEntry := request.(AppendEntry)
	ra.appendEntryChan <- appendEntry
}

//Process will apply the appendEntry logic
func (ra *raftAppendEntry) Process(meta RaftRpcMeta) (RaftRpcResponse, error) {
	log.Printf("Processing AppendEntry %v\n", meta)
	appendEntryMeta := meta.(AppendEntryMeta)
	currentTerm := ra.raftTerm.GetTerm()
	// 1. Reply false if term < currentTerm (§5.1)
	if appendEntryMeta.Term < currentTerm {
		return models.AppendEntryResponse{
			Term:    appendEntryMeta.Term,
			Success: false,
		}, nil
	}
	// 2. Reply false if log doesn’t contain an entry at prevLogIndex
	//    whose term matches prevLogTerm (§5.3)
	prevLogIndex := appendEntryMeta.PrevLogIndex
	em, err := ra.raftLog.GetLogEntryMetaAtIndex(prevLogIndex)
	if err != nil {
		log.Printf("Encountered error while querying log for index")
		return nil, err
	}
	if int64(em.Term) != appendEntryMeta.PrevLogTerm {
		// 3. If an existing entry conflicts with a new one (same index
		//    but different terms), delete the existing entry and all that
		//    follow it (§5.3)
		err := ra.raftLog.TruncateFromIndex(prevLogIndex)
		if err != nil {
			log.Printf("Encountered error while truncating log")
		}
		return models.AppendEntryResponse{Term: appendEntryMeta.Term, Success: false}, nil
	}
	// 4. Append any new entries not already in the log
	resp, err := ra.raftLog.AppendEntry(raftlog.Entry{
		Term:    uint64(appendEntryMeta.Term),
		Payload: appendEntryMeta.Entries,
	})
	// 5. If leaderCommit > commitIndex, set commitIndex =
	//     min(leaderCommit, index of last new entry)
	leaderCommit := appendEntryMeta.LeaderCommit
	if leaderCommit > ra.raftIndex.GetCommitOffset() {
		if leaderCommit > resp.LogIndex {
			ra.raftIndex.SetCommitOffset(resp.LogIndex)
		}
		ra.raftIndex.SetCommitOffset(leaderCommit)
	}
	return models.AppendEntryResponse{Term: appendEntryMeta.Term, Success: true}, nil
}

func (ra *raftAppendEntry) RaftRpcReqChan() <-chan RaftRpcRequest {
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

type RequestVoteResponse struct {
	Term        int64
	VoteGranted bool
}

func NewRaftRequestVote(raftTerm term.RaftTerm,
	raftLog raftlog.RaftLog) RaftRpc {
	requestVoteChan := make(chan RaftRpcRequest, 1)
	return raftRequestVote{
		requestVoteChan: requestVoteChan,
		raftTerm:        raftTerm,
		raftLog:         raftLog,
	}
}

type raftRequestVote struct {
	requestVoteChan chan RaftRpcRequest
	raftTerm        term.RaftTerm
	raftLog         raftlog.RaftLog
	votedFor        string
}

func (rrV raftRequestVote) Receive(request RaftRpcRequest) {
	requestVote := request.(RequestVote)
	rrV.requestVoteChan <- requestVote
}

/*
Process
  1. Reply false if term < currentTerm (§5.1)
  2. If votedFor is null or candidateId, and candidate’s log is at
      least as up-to-date as receiver’s log, grant vote (§5.2, §5.4)
*/
func (rrV raftRequestVote) Process(meta RaftRpcMeta) (RaftRpcResponse, error) {
	log.Println("Received request vote", meta)
	requestVoteMeta := meta.(RequestVoteMeta)
	currentTerm := rrV.raftTerm.GetTerm()
	if requestVoteMeta.Term < currentTerm {
		return RequestVoteResponse{Term: currentTerm, VoteGranted: false}, nil
	}
	votedFor := rrV.votedFor
	if (votedFor == "" || votedFor == requestVoteMeta.CandidateId) &&
		rrV.isLogUptoDate(requestVoteMeta) {
		rrV.votedFor = requestVoteMeta.CandidateId
		return RequestVoteResponse{
			Term:        requestVoteMeta.Term,
			VoteGranted: true,
		}, nil
	}
	return RequestVoteResponse{Term: currentTerm, VoteGranted: false}, nil
}

func (rrV raftRequestVote) RaftRpcReqChan() <-chan RaftRpcRequest {
	return rrV.requestVoteChan
}

func (rrV raftRequestVote) ReceiveRequestVote(requestVote RequestVote) {
	rrV.requestVoteChan <- requestVote
}

//Raft determines which of two logs is more up-to-date
//by comparing the index and term of the last entries in the
//logs. If the logs have last entries with different terms, then
//the log with the later term is more up-to-date. If the logs
//end with the same term, then whichever log is longer is
//more up-to-date.
func (rrV raftRequestVote) isLogUptoDate(meta RequestVoteMeta) bool {
	entryMeta := rrV.raftLog.GetCurrentLogEntry()
	currentTerm := int64(entryMeta.Term)
	currentIndex := int64(entryMeta.Index)
	if currentTerm < meta.LastLogTerm {
		return true
	}
	if currentTerm == meta.LastLogTerm {
		return meta.LastLogIndex > currentIndex
	}
	return false
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
	raftterm := rcc.raftTerm.GetTerm()
	leader := rcc.raftMember.Leader()
	payload := clientCommandMeta.Payload
	aer, err := rcc.raftLog.AppendEntry(raftlog.Entry{
		Term:    uint64(raftterm),
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
	for _, m := range members {
		entry := appendentry.Entry{
			MemberID:     m.ID,
			RespChan:     respChan,
			Term:         raftterm,
			LeaderID:     leader.ID,
			PrevLogIndex: aer.PrevLogIndex,
			PrevLogTerm:  aer.PrevLogTerm,
			Entries:      payload,
			LeaderCommit: rcc.raftIndex.GetCommitOffset(),
			MemberAddr:   m.Address,
		}
		rcc.appendEntrySender.ForwardEntry(entry)
	}
	if err = rcc.waitForMajorityAcks(len(members), 100*time.Millisecond, respChan); err != nil {
		return ClientCommandResponse{Committed: false}, nil
	}
	//If majority votes are received, set the commit index
	rcc.raftIndex.SetCommitOffset(aer.LogIndex)
	//Apply the command
	if err = rcc.raftApplicator.Apply(payload); err == nil {
		//Upon successful application, set the applyIndex
		rcc.raftIndex.SetApplyOffset(aer.LogIndex)
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
