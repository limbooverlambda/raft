package actors

import (
	"github.com/kitengo/raft/internal/models"
	"log"
	"time"

	"github.com/kitengo/raft/internal/locator"
	raftconfig "github.com/kitengo/raft/internal/rconfig"
	raftrpc "github.com/kitengo/raft/internal/rpc"
	raftstate "github.com/kitengo/raft/internal/state"
	raftterm "github.com/kitengo/raft/internal/term"
	rafttimer "github.com/kitengo/raft/internal/timer"
	raftvoter "github.com/kitengo/raft/internal/voter"
)

type Candidate interface {
	Run()
}

type candidate struct {
	state     raftstate.RaftState
	term      raftterm.RaftTerm
	voter     raftvoter.RaftVoter
	raftTimer rafttimer.RaftTimer
	aeRPC     raftrpc.RaftRpc
	voteRPC   raftrpc.RaftRpc
}

// Run On conversion to candidate,
// Start election:
//• Increment currentTerm
//• Vote for self
//• Reset election timer
//• Send RequestVote RPCs to all other servers
//• If votes received from majority of servers: become leader
//• If AppendEntries RPC received from new leader: convert to
//  follower
//• If election timeout elapses: start new election
func (c *candidate) Run() {
	log.Println("Candidate")
	aeReqChan := c.aeRPC.RaftRpcReqChan()
	voteReqChan := c.voteRPC.RaftRpcReqChan()
	term := c.term.IncrementTerm()
	reqVoteChan := c.voter.RequestVote(term)
	//Start election timer
	c.raftTimer.SetDeadline(time.Now())
	ticker := time.NewTicker(raftconfig.PollDuration)
	defer ticker.Stop()
	for tick := range ticker.C {
		select {
		case vs := <-reqVoteChan:
			{
				var state raftstate.State
				if vs == raftvoter.Follower {
					state = raftstate.FollowerState
				}
				if vs == raftvoter.Leader {
					state = raftstate.LeaderState
				}
				c.state.SetState(state)
				return

			}
		case aeReq := <-aeReqChan:
			{
				respChan, errChan := aeReq.GetResponseChan(), aeReq.GetErrorChan()
				aer := aeReq.(raftrpc.AppendEntry)
				resp, err := c.aeRPC.Process(raftrpc.AppendEntryMeta{
					Term:         aer.Term,
					LeaderId:     aer.LeaderId,
					PrevLogIndex: aer.PrevLogIndex,
					PrevLogTerm:  aer.PrevLogTerm,
					Entries:      aer.Entries,
					LeaderCommit: aer.LeaderCommit,
				})
				if err != nil {
					errChan <- err
				}
				respChan <- resp
				//reset the deadline
				aeResp := resp.(models.AppendEntryResponse)
				if aeResp.Term > term {
					c.state.SetState(raftstate.FollowerState)
					return
				}
			}
		case voteReq := <-voteReqChan:
			{
				respChan, errChan := voteReq.GetResponseChan(), voteReq.GetErrorChan()
				vr := voteReq.(raftrpc.RequestVote)
				resp, err := c.voteRPC.Process(raftrpc.RequestVoteMeta{
					Term:         vr.Term,
					CandidateId:  vr.CandidateId,
					LastLogIndex: vr.LastLogIndex,
					LastLogTerm:  vr.LastLogTerm,
				})
				if err != nil {
					errChan <- err
				}
				respChan <- resp
			}
		default:
			if tick.After(c.raftTimer.GetDeadline()) {
				term = c.term.IncrementTerm()
				reqVoteChan = c.voter.RequestVote(term)
			}
		}
	}

}

type CandidateProvider interface {
	Provide() Candidate
}

type candidateProvider struct {
	locator.ServiceLocator
}

func (cp *candidateProvider) Provide() Candidate {
	rpcLocator := cp.GetRpcLocator()
	return &candidate{
		state:     cp.GetRaftState(),
		term:      cp.GetRaftTerm(),
		voter:     cp.GetRaftVoter(),
		raftTimer: cp.GetRaftTimer(),
		aeRPC:     rpcLocator.GetAppendEntrySvc(),
		voteRPC:   rpcLocator.GetRequestVoteSvc(),
	}
}
