package actors

import (
	"log"
	"time"

	svclocator "github.com/kitengo/raft/internal/locator"
	raftconfig "github.com/kitengo/raft/internal/rconfig"
	raftrpc "github.com/kitengo/raft/internal/rpc"
	raftstate "github.com/kitengo/raft/internal/state"
	rafttimer "github.com/kitengo/raft/internal/timer"
)

type Follower interface {
	Run()
}

type follower struct {
	state     raftstate.RaftState
	aeRPC     raftrpc.RaftRpc
	voteRPC   raftrpc.RaftRpc
	raftTimer rafttimer.RaftTimer
}

func (f *follower) Run() {
	log.Println("Follower")
	aeReqChan := f.aeRPC.RaftRpcReqChan()
	voteReqChan := f.voteRPC.RaftRpcReqChan()
	f.raftTimer.SetDeadline(time.Now())
	ticker := time.NewTicker(raftconfig.PollDuration)
	defer ticker.Stop()
	for tick := range ticker.C {
		select {
		case aeReq := <-aeReqChan:
			{
				respChan, errChan := aeReq.GetResponseChan(), aeReq.GetErrorChan()
				aer := aeReq.(raftrpc.AppendEntry)
				resp, err := f.aeRPC.Process(raftrpc.AppendEntryMeta{
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
				f.raftTimer.SetDeadline(tick)
			}
		case voteReq := <-voteReqChan:
			{
				respChan, errChan := voteReq.GetResponseChan(), voteReq.GetErrorChan()
				vr := voteReq.(raftrpc.RequestVote)
				resp, err := f.voteRPC.Process(raftrpc.RequestVoteMeta{
					Term:         vr.Term,
					CandidateId:  vr.CandidateId,
					LastLogIndex: vr.LastLogIndex,
					LastLogTerm:  vr.LastLogTerm,
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
				f.state.SetState(raftstate.CandidateState)
				return
			}
		}
	}
}

type FollowerProvider interface {
	Provide() Follower
}

type followerProvider struct {
	svclocator.ServiceLocator
}

func (fp *followerProvider) Provide() Follower {
	rpcLocator := fp.GetRpcLocator()
	return &follower{
		state:     fp.GetRaftState(),
		aeRPC:     rpcLocator.GetAppendEntrySvc(),
		voteRPC:   rpcLocator.GetRequestVoteSvc(),
		raftTimer: fp.GetRaftTimer(),
	}
}
