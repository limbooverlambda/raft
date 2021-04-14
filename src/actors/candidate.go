package actors

import (
	"log"
	"time"

	"kitengo/raft/rconfig"
	"kitengo/raft/rpc"
	"kitengo/raft/state"
	"kitengo/raft/term"
	"kitengo/raft/timer"
	"kitengo/raft/voter"
)

type Candidate interface {
	Run()
}

type candidate struct{
	state state.RaftState
	term term.RaftTerm
	voter voter.RaftVoter
	raftTimer timer.RaftTimer
	aeRPC rpc.RaftAppendEntry
	voteRPC rpc.RaftRequestVote
}

//On conversion to candidate, start election:
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
	c.raftTimer.SetDeadline(time.Now())
	ticker := time.NewTicker(rconfig.PollDuration)
	defer ticker.Stop()
	for tick := range ticker.C {
		select {
		case vs := <- reqVoteChan:
			{
				if vs == voter.Leader {
					c.state.SetState(state.LeaderState)
					return
				}
			}
		case aeReq := <-aeReqChan:{
			respChan, errChan := aeReq.RespChan, aeReq.ErrorChan
			resp, err := c.aeRPC.Process(rpc.AppendEntryMeta{
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
				c.state.SetState(state.FollowerState)
				return
			}
		}
		case voteReq := <-voteReqChan:{
			respChan, errChan := voteReq.RespChan, voteReq.ErrorChan
			resp, err := c.voteRPC.Process(rpc.RequestVoteMeta{
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
	Provide(raftState state.RaftState) Candidate
}

type candidateProvider struct{}

func (candidateProvider) Provide(raftState state.RaftState) Candidate {
	return &candidate{state: raftState}
}
