package actors

import (
	"context"
	"log"
	"time"

	svclocator "github.com/kitengo/raft/internal/locator"
	raftconfig "github.com/kitengo/raft/internal/rconfig"
	raftrpc "github.com/kitengo/raft/internal/rpc"
	raftstate "github.com/kitengo/raft/internal/state"
	rafttimer "github.com/kitengo/raft/internal/timer"
)

type Follower interface {
	Run(ctx context.Context)
}

type follower struct {
	state     raftstate.RaftState
	aeRPC     raftrpc.RaftRpc
	voteRPC   raftrpc.RaftRpc
	raftTimer rafttimer.RaftTimer
}

func (f *follower) Run(ctx context.Context) {
	log.Println("Setting state to Follower")
	f.state.SetState(raftstate.FollowerState)
	aeReqChan := f.aeRPC.RaftRpcReqChan()
	voteReqChan := f.voteRPC.RaftRpcReqChan()
	f.raftTimer.SetDeadline(time.Now())
	ticker := time.NewTicker(raftconfig.PollDuration)
	defer ticker.Stop()
	for tick := range ticker.C {
		select {
		case <-ctx.Done():
			{
				log.Println("Cancelling the running follower")
				return
			}
		case aeReq := <-aeReqChan:
			{
				respChan, errChan := aeReq.GetResponseChan(), aeReq.GetErrorChan()
				aer := aeReq.(raftrpc.AppendEntry)
				aeMeta := raftrpc.AppendEntryMeta{
					Term:         aer.Term,
					LeaderId:     aer.LeaderId,
					PrevLogIndex: aer.PrevLogIndex,
					PrevLogTerm:  aer.PrevLogTerm,
					Entries:      aer.Entries,
					LeaderCommit: aer.LeaderCommit,
				}
				resp, err := f.aeRPC.Process(aeMeta)
				if err != nil {
					errChan <- err
				} else {
					respChan <- resp
				}
				//reset the deadline
				f.raftTimer.SetDeadline(tick)
			}
		case voteReq := <-voteReqChan:
			{
				respChan, errChan := voteReq.GetResponseChan(), voteReq.GetErrorChan()
				vr := voteReq.(raftrpc.RequestVote)
				reqVoteMeta := raftrpc.RequestVoteMeta{
					Term:         vr.Term,
					CandidateId:  vr.CandidateId,
					LastLogIndex: vr.LastLogIndex,
					LastLogTerm:  vr.LastLogTerm,
				}
				resp, err := f.voteRPC.Process(reqVoteMeta)
				if err != nil {
					errChan <- err
				} else {
					respChan <- resp
				}
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
