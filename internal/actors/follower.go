package actors

import (
	"log"
	"time"

	"github.com/kitengo/raft/internal/rconfig"
	"github.com/kitengo/raft/internal/rpc"
	"github.com/kitengo/raft/internal/state"
	"github.com/kitengo/raft/internal/timer"
)

type Follower interface {
	Run()
}

type follower struct{
	state     state.RaftState
	aeRPC     rpc.RaftAppendEntry
	voteRPC   rpc.RaftRequestVote
	raftTimer timer.RaftTimer
}

func (f *follower) Run() {
	log.Println("Follower")
	aeReqChan := f.aeRPC.AppendEntryReqChan()
	voteReqChan := f.voteRPC.RequestVoteReqChan()
	f.raftTimer.SetDeadline(time.Now())
	ticker := time.NewTicker(rconfig.PollDuration)
	defer ticker.Stop()
	for tick := range ticker.C {
		select {
		case aeReq := <-aeReqChan:{
			respChan, errChan := aeReq.RespChan, aeReq.ErrorChan
			resp, err := f.aeRPC.Process(rpc.AppendEntryMeta{
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
			f.raftTimer.SetDeadline(tick)
		}
		case voteReq := <-voteReqChan:{
			respChan, errChan := voteReq.RespChan, voteReq.ErrorChan
			resp, err := f.voteRPC.Process(rpc.RequestVoteMeta{
				Term:         voteReq.Term,
				CandidateId:  voteReq.CandidateId,
				LastLogIndex: voteReq.LastLogIndex,
				LastLogTerm:  voteReq.LastLogTerm,
			})
			if err != nil {
				errChan <- err
			}
			respChan <- resp
			//reset the deadline
			f.raftTimer.SetDeadline(tick)
		}
		default:
			if tick.After(f.raftTimer.GetDeadline()) {
				f.state.SetState(state.CandidateState)
				return
			}
		}
	}
}

type FollowerProvider interface {
	Provide(raftState state.RaftState) Follower
}

type followerProvider struct{}

func (followerProvider) Provide(raftState state.RaftState) Follower {
	return &follower{state: raftState}
}


