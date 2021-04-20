package actors

import (
	"log"
	"time"

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

type candidate struct{
	state     raftstate.RaftState
	term      raftterm.RaftTerm
	voter     raftvoter.RaftVoter
	raftTimer rafttimer.RaftTimer
	aeRPC     raftrpc.RaftAppendEntry
	voteRPC   raftrpc.RaftRequestVote
}

//On conversion to candidate,
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
	aeReqChan := c.aeRPC.AppendEntryReqChan()
	voteReqChan := c.voteRPC.RequestVoteReqChan()
	term := c.term.IncrementTerm()
	reqVoteChan := c.voter.RequestVote(term)
	//Start election timer
	c.raftTimer.SetDeadline(time.Now())
	ticker := time.NewTicker(raftconfig.PollDuration)
	defer ticker.Stop()
	for tick := range ticker.C {
		select {
		case vs := <- reqVoteChan:
			{
				if vs == raftvoter.Leader {
					c.state.SetState(raftstate.LeaderState)
					return
				}
			}
		case aeReq := <-aeReqChan:{
			respChan, errChan := aeReq.RespChan, aeReq.ErrorChan
			resp, err := c.aeRPC.Process(raftrpc.AppendEntryMeta{
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
			//reset the deadline
			if resp.Term > term {
				c.state.SetState(raftstate.FollowerState)
				return
			}
		}
		case voteReq := <-voteReqChan:{
			respChan, errChan := voteReq.RespChan, voteReq.ErrorChan
			resp, err := c.voteRPC.Process(raftrpc.RequestVoteMeta{
				Term:         voteReq.Term,
				CandidateId:  voteReq.CandidateId,
				LastLogIndex: voteReq.LastLogIndex,
				LastLogTerm:  voteReq.LastLogTerm,
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
	Provide(raftState raftstate.RaftState) Candidate
}

type candidateProvider struct{}

func (candidateProvider) Provide(raftState raftstate.RaftState) Candidate {
	return &candidate{state: raftState}
}
