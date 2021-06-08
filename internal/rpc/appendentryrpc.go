package rpc

import (
	raftlog "github.com/kitengo/raft/internal/log"
	"github.com/kitengo/raft/internal/models"
	raftstate "github.com/kitengo/raft/internal/state"
	"github.com/kitengo/raft/internal/term"
	"log"
)

type AppendEntry struct {
	Term         uint64
	LeaderId     string
	PrevLogIndex uint64
	PrevLogTerm  uint64
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
	Term         uint64
	LeaderId     string
	PrevLogIndex uint64
	PrevLogTerm  uint64
	Entries      []byte
	LeaderCommit uint64
}

func NewRaftAppendEntry(raftTerm term.RaftTerm,
	raftLog raftlog.RaftLog, raftIndex *raftstate.RaftIndex) RaftRpc {
	appendEntryChan := make(chan RaftRpcRequest, 1)
	return &raftAppendEntry{
		raftTerm:        raftTerm,
		raftLog:         raftLog,
		raftIndex:       raftIndex,
		appendEntryChan: appendEntryChan,
	}
}

type raftAppendEntry struct {
	raftTerm        term.RaftTerm
	raftLog         raftlog.RaftLog
	raftIndex       *raftstate.RaftIndex
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
	//If entry is a heartbeat (payload size is 0), return the success quickly
	if len(appendEntryMeta.Entries) == 0 {
		log.Printf("Received heartbeat...")
		return models.AppendEntryResponse{
			Term:    appendEntryMeta.Term,
			Success: true,
		}, nil
	}

	// 2. Reply false if log doesn’t contain an entry at prevLogIndex
	//    whose term matches prevLogTerm (§5.3)
	incomingLogIndex := appendEntryMeta.PrevLogIndex
	logEntryMeta, err := ra.raftLog.LogEntryMeta(incomingLogIndex)
	if err != nil {
		log.Printf("Encountered error while querying log for index %d: %v\n", incomingLogIndex, err)
		return nil, err
	}

	incomingLogTerm := appendEntryMeta.Term
	currentLogTerm := logEntryMeta.Term
	if incomingLogTerm != currentLogTerm {
		log.Printf("Log mismatch incomingLogTerm %d currentLogTerm %d\n", incomingLogTerm, currentLogTerm)
		em, err := ra.raftLog.LogEntryMeta(incomingLogIndex)
		if err != nil {
			log.Printf("Encountered error while querying log for index %d: %v\n", incomingLogIndex, err)
			return nil, err
		}
		if em.Term != appendEntryMeta.PrevLogTerm {
			// 3. If an existing entry conflicts with a new one (same index
			//    but different terms), delete the existing entry and all that
			//    follow it (§5.3)
			err := ra.raftLog.Truncate(incomingLogIndex)
			if err != nil {
				log.Printf("Encountered error while truncating log")
			}
			return models.AppendEntryResponse{Term: appendEntryMeta.Term, Success: false}, nil
		}
	}
	// 4. Append any new entries not already in the log
	resp, err := ra.raftLog.AppendEntry(raftlog.Entry{
		Term:    uint64(appendEntryMeta.Term),
		Payload: appendEntryMeta.Entries,
	})
	if err != nil {
		log.Printf("unable to append entry to log due to %v\n", err)
		return nil, err
	}
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
