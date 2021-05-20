package actors

import (
	"github.com/kitengo/raft/internal/models"
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
	clientRPC raftrpc.RaftRpc
	aeRPC     raftrpc.RaftRpc
	voteRPC   raftrpc.RaftRpc
	raftTerm  raftterm.RaftTerm
	raftTimer rafttimer.RaftTimer
}

func (l *leader) Run() {
	log.Println("Leader")
	clientReqChan := l.clientRPC.RaftRpcReqChan()
	aeReqChan := l.aeRPC.RaftRpcReqChan()
	voteReqChan := l.voteRPC.RaftRpcReqChan()
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
				respChan, errChan := clientReq.GetResponseChan(), clientReq.GetErrorChan()
				clReq := clientReq.(raftrpc.ClientCommand)
				resp, err := l.clientRPC.Process(raftrpc.ClientCommandMeta{
					Payload: clReq.Payload,
				})
				l.raftTimer.SetIdleTimeout()
				if err != nil {
					errChan <- err
				}
				respChan <- resp
			}
		case aeReq := <-aeReqChan:
			{
				currentTerm := l.raftTerm.GetTerm()
				respChan, errChan := aeReq.GetResponseChan(), aeReq.GetErrorChan()
				aer := aeReq.(raftrpc.AppendEntry)
				resp, err := l.aeRPC.Process(raftrpc.AppendEntryMeta{
					Term:         aer.Term,
					LeaderId:     aer.LeaderId,
					PrevLogIndex: aer.PrevLogIndex,
					PrevLogTerm:  aer.PrevLogTerm,
					Entries:      aer.Entries,
					LeaderCommit: aer.LeaderCommit,
				})
				if err != nil {
					errChan <- err
				} else {
					respChan <- resp
					aeResp := resp.(models.AppendEntryResponse)
					if aeResp.Term > currentTerm {
						log.Println("Received a higher term. flipping back to be a follower")
						l.state.SetState(raftstate.FollowerState)
						return
					}
				}
			}
		case voteReq := <-voteReqChan:
			{
				currentTerm := l.raftTerm.GetTerm()
				respChan, errChan := voteReq.GetResponseChan(), voteReq.GetErrorChan()
				vr := voteReq.(raftrpc.RequestVote)
				resp, err := l.voteRPC.Process(raftrpc.RequestVoteMeta{
					Term:         vr.Term,
					CandidateId:  vr.CandidateId,
					LastLogIndex: vr.LastLogIndex,
					LastLogTerm:  vr.LastLogTerm,
				})
				if err != nil {
					errChan <- err
				}
				respChan <- resp
				vResp := resp.(raftrpc.RequestVoteResponse)
				if vResp.Term > currentTerm {
					log.Println("Received a higher term. flipping back to be a follower")
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

type leaderProvider struct {
	svclocator.ServiceLocator
}

func (lp *leaderProvider) Provide() Leader {
	rpcLocator := lp.GetRpcLocator()
	return &leader{
		state:     lp.GetRaftState(),
		heartbeat: lp.GetRaftHeartbeat(),
		clientRPC: rpcLocator.GetClientCommandSvc(),
		aeRPC:     rpcLocator.GetAppendEntrySvc(),
		voteRPC:   rpcLocator.GetRequestVoteSvc(),
		raftTerm:  lp.GetRaftTerm(),
		raftTimer: lp.GetRaftTimer(),
	}
}
