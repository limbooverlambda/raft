package actors

import (
	"log"
	"time"

	"github.com/kitengo/raft/internal/heartbeat"
	"github.com/kitengo/raft/internal/rconfig"
	"github.com/kitengo/raft/internal/rpc"
	"github.com/kitengo/raft/internal/state"
	"github.com/kitengo/raft/internal/term"
	"github.com/kitengo/raft/internal/timer"
)

type Leader interface {
	Run()
}

type leader struct {
	state     state.RaftState
	heartbeat heartbeat.RaftHeartbeat
	clientRPC rpc.RaftClientCommand
	aeRPC     rpc.RaftAppendEntry
	voteRPC   rpc.RaftRequestVote
	raftTerm  term.RaftTerm
	raftTimer timer.RaftTimer
}

func (l *leader) Run() {
	log.Println("Leader")
	clientReqChan := l.clientRPC.ClientCommandReqChan()
	aeReqChan := l.aeRPC.AppendEntryReqChan()
	voteReqChan := l.voteRPC.RequestVoteReqChan()
	//Upon election: send initial empty AppendEntries RPCs (heartbeat)
	l.heartbeat.SendHeartbeats()
	//Reset the idle time
	l.raftTimer.SetIdleTimeout()
	//Start the ticker
	ticker := time.NewTicker(rconfig.PollDuration)
	defer ticker.Stop()
	for tick := range ticker.C {
		select {
		// If command received from client: append entry to local log,
		// respond after entry applied to state machine (§5.3)
		case clientReq := <-clientReqChan:
			{
				respChan, errChan := clientReq.RespChan, clientReq.ErrorChan
				resp, err := l.clientRPC.Process(rpc.ClientCommandMeta{
					Payload: clientReq.Payload,
				})
				l.raftTimer.SetIdleTimeout()
				if err != nil {
					errChan <- err
					return
				}
				respChan <- resp
				return
			}
		case aeReq := <-aeReqChan:
			{
				currentTerm := l.raftTerm.GetTerm()
				respChan, errChan := aeReq.RespChan, aeReq.ErrorChan
				resp, err := l.aeRPC.Process(rpc.AppendEntryMeta{
					Term:         aeReq.Term,
					LeaderId:     aeReq.LeaderId,
					PrevLogIndex: aeReq.PrevLogIndex,
					PrevLogTerm:  aeReq.PrevLogTerm,
					Entries:      aeReq.Entries,
					LeaderCommit: aeReq.LeaderCommit,
				})
				if err != nil {
					errChan <- err
				}
				respChan <- resp
				if resp.Term > currentTerm {
					l.state.SetState(state.FollowerState)
					return
				}
			}
		case voteReq := <-voteReqChan:
			{
				currentTerm := l.raftTerm.GetTerm()
				respChan, errChan := voteReq.RespChan, voteReq.ErrorChan
				resp, err := l.voteRPC.Process(rpc.RequestVoteMeta{
					Term:         voteReq.Term,
					CandidateId:  voteReq.CandidateId,
					LastLogIndex: voteReq.LastLogIndex,
					LastLogTerm:  voteReq.LastLogTerm,
				})
				if err != nil {
					errChan <- err
				}
				respChan <- resp
				if resp.Term > currentTerm {
					l.state.SetState(state.FollowerState)
					return
				}
			}
		default:
			if tick.After(l.raftTimer.GetIdleTimeout()) {
				l.heartbeat.SendHeartbeats()
			}
		}
	}
}

type LeaderProvider interface {
	Provide(raftState state.RaftState) Leader
}

type leaderProvider struct{}

func (leaderProvider) Provide(raftState state.RaftState) Leader {
	return &leader{state: raftState}
}
