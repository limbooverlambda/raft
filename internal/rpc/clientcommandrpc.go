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
			MemberPort:   m.Port,
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
