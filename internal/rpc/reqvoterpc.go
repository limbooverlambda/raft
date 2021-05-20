package rpc

import (
	"errors"
	raftlog "github.com/kitengo/raft/internal/log"
	"github.com/kitengo/raft/internal/member"
	"github.com/kitengo/raft/internal/term"
	"log"
)

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
	raftLog raftlog.RaftLog,
	raftMember member.RaftMember) RaftRpc {
	requestVoteChan := make(chan RaftRpcRequest, 1)
	return raftRequestVote{
		requestVoteChan: requestVoteChan,
		raftTerm:        raftTerm,
		raftLog:         raftLog,
		raftMember:      raftMember,
	}
}

type raftRequestVote struct {
	requestVoteChan chan RaftRpcRequest
	raftTerm        term.RaftTerm
	raftLog         raftlog.RaftLog
	raftMember      member.RaftMember
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
	requestVoteMeta, ok := meta.(RequestVoteMeta)
	if !ok {
		return RequestVoteResponse{}, errors.New("failed to receive request vote")
	}
	currentTerm := rrV.raftTerm.GetTerm()
	if requestVoteMeta.Term < currentTerm {
		log.Println("Current term is greater than incoming term, rejecting vote request from", requestVoteMeta.CandidateId)
		return RequestVoteResponse{Term: currentTerm, VoteGranted: false}, nil
	}
	votedFor := rrV.raftMember.VotedFor()
	isLogUptoDate := rrV.isLogUptoDate(requestVoteMeta)
	didIVote := votedFor == "" || votedFor == requestVoteMeta.CandidateId
	if didIVote && isLogUptoDate {
		log.Printf("Accepting the vote from %+v\n", requestVoteMeta)
		rrV.raftMember.SetVotedFor(requestVoteMeta.CandidateId)
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
		return meta.LastLogIndex >= currentIndex
	}
	return false
}
