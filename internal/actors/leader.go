package actors

import (
	"log"
	"time"

	raftheartbeat "github.com/kitengo/raft/internal/heartbeat"
	svclocator "github.com/kitengo/raft/internal/locator"
	raftconfig "github.com/kitengo/raft/internal/rconfig"
	raftrpc "github.com/kitengo/raft/internal/rpc"
	raftstate "github.com/kitengo/raft/internal/state"
	raftterm "github.com/kitengo/raft/internal/term"
	rafttimer "github.com/kitengo/raft/internal/timer"
)

type Leader interface {
	Run()
}

type leader struct {
	state     raftstate.RaftState
	heartbeat raftheartbeat.RaftHeartbeat
	clientRPC raftrpc.RaftClientCommand
	aeRPC     raftrpc.RaftAppendEntry
	voteRPC   raftrpc.RaftRequestVote
	raftTerm  raftterm.RaftTerm
	raftTimer rafttimer.RaftTimer
}

func (l *leader) Run() {
	log.Println("Leader")
	clientReqChan := l.clientRPC.ClientCommandReqChan()
	aeReqChan := l.aeRPC.AppendEntryReqChan()
	voteReqChan := l.voteRPC.RequestVoteReqChan()
	//Upon election: send initial empty AppendEntries RPCs (raftheartbeat)
	l.heartbeat.SendHeartbeats()
	//Reset the idle time
	l.raftTimer.SetIdleTimeout()
	//Start the ticker
	ticker := time.NewTicker(raftconfig.PollDuration)
	defer ticker.Stop()
	for tick := range ticker.C {
		select {
		// If command received from client: append entry to local log,
		// respond after entry applied to raftstate machine (§5.3)
		case clientReq := <-clientReqChan:
			{
				respChan, errChan := clientReq.RespChan, clientReq.ErrorChan
				resp, err := l.clientRPC.Process(raftrpc.ClientCommandMeta{
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
				resp, err := l.aeRPC.Process(raftrpc.AppendEntryMeta{
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
					l.state.SetState(raftstate.FollowerState)
					return
				}
			}
		case voteReq := <-voteReqChan:
			{
				currentTerm := l.raftTerm.GetTerm()
				respChan, errChan := voteReq.RespChan, voteReq.ErrorChan
				resp, err := l.voteRPC.Process(raftrpc.RequestVoteMeta{
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
					l.state.SetState(raftstate.FollowerState)
					return
				}
			}
		default:
			if tick.After(l.raftTimer.GetIdleTimeout()) {
				//Idle expired, send the heartbeat
				l.heartbeat.SendHeartbeats()
				//Reset the timeout
				l.raftTimer.SetIdleTimeout()
			}
		}
	}
}

type LeaderProvider interface {
	Provide() Leader
}

type leaderProvider struct{
	svclocator.ServiceLocator
}

func (lp *leaderProvider) Provide() Leader {
	return &leader{
		state:     lp.GetRaftState(),
		heartbeat: lp.GetRaftHeartbeat(),
		clientRPC: lp.GetRpcLocator().GetClientCommandSvc(),
		aeRPC:     lp.GetRpcLocator().GetAppendEntrySvc(),
		voteRPC:   lp.GetRpcLocator().GetRequestVoteSvc(),
		raftTerm:  lp.GetRaftTerm(),
		raftTimer: lp.GetRaftTimer(),
	}
}


