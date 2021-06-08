package actors

import (
	"context"
	"github.com/kitengo/raft/internal/member"
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
	Run(ctx context.Context)
}

type candidate struct {
	member    member.RaftMember
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
func (c *candidate) Run(ctx context.Context) {
	log.Println("Setting state to Candidate")
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
		case <-ctx.Done():
			{
				log.Println("Cancelling the running candidate")
				return
			}
		case vs := <-reqVoteChan:
			{
				state := c.processVoteStatus(vs)
				log.Println("Candidate, voteStatus", state)
				c.state.SetState(state)
				switch state {
				case raftstate.FollowerState:
					log.Println("candidate -> follower")
					return
				case raftstate.LeaderState:
					log.Println("candidate ->  leader")
					myID := c.member.Self().ID
					c.member.SetLeaderID(myID)
					return
				default:
					log.Println("no majority, staying behind as ", state)
				}
			}
		case aeReq := <-aeReqChan:
			{
				if respTerm, respLeaderID, err := c.processAERequest(aeReq); err == nil {
					if respTerm >= term {
						if respLeaderID != "" {
							c.member.SetLeaderID(respLeaderID)
						}
						c.state.SetState(raftstate.FollowerState)
						return
					}
				}
			}
		case voteReq := <-voteReqChan:
			{
				c.processVoteReq(voteReq)
			}
		default:
			reqVoteChan = c.checkForExceededDeadline(tick, term)
		}
	}

}

func (c *candidate) checkForExceededDeadline(tick time.Time, term uint64) <-chan raftvoter.VoteStatus {
	if tick.After(c.raftTimer.GetDeadline()) {
		term = c.term.IncrementTerm()
		reqVoteChan := c.voter.RequestVote(term)
		c.raftTimer.SetDeadline(time.Now())
		return reqVoteChan
	}
	return nil
}

func (c *candidate) processVoteReq(voteReq raftrpc.RaftRpcRequest) {
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
	} else {
		respChan <- resp
	}
}

func (c *candidate) processAERequest(req raftrpc.RaftRpcRequest) (term uint64, leaderID string, err error) {
	respChan, errChan := req.GetResponseChan(), req.GetErrorChan()
	aeReq := req.(raftrpc.AppendEntry)
	aeMeta := raftrpc.AppendEntryMeta{
		Term:         aeReq.Term,
		LeaderId:     aeReq.LeaderId,
		PrevLogIndex: aeReq.PrevLogIndex,
		PrevLogTerm:  aeReq.PrevLogTerm,
		Entries:      aeReq.Entries,
		LeaderCommit: aeReq.LeaderCommit,
	}
	resp, err := c.aeRPC.Process(aeMeta)
	if err != nil {
		errChan <- err
		return
	} else {
		respChan <- resp
		aeResp := resp.(models.AppendEntryResponse)
		term = aeResp.Term
		leaderID = aeReq.LeaderId
		return
	}
}

func (c *candidate) processVoteStatus(vs raftvoter.VoteStatus) raftstate.State {
	switch vs {
	case raftvoter.Follower:
		return raftstate.FollowerState
	case raftvoter.Leader:
		return raftstate.LeaderState
	default:
		return raftstate.CandidateState
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
		member:    cp.GetRaftMember(),
		aeRPC:     rpcLocator.GetAppendEntrySvc(),
		voteRPC:   rpcLocator.GetRequestVoteSvc(),
	}
}
