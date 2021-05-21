package actors

import (
	"github.com/kitengo/raft/internal/member"
	raftrpc "github.com/kitengo/raft/internal/rpc"
	raftstate "github.com/kitengo/raft/internal/state"
	raftterm "github.com/kitengo/raft/internal/term"
	rafttimer "github.com/kitengo/raft/internal/timer"
	raftvoter "github.com/kitengo/raft/internal/voter"
	"time"
)

type candidateStub struct {
	fakeMember  fakeMember
	fakeState   fakeRaftState
	fakeTerm    fakeRaftTerm
	fakeVoter   fakeRaftVoter
	fakeTimer   fakeRaftTimer
	fakeAeRPC   fakeAeRPC
	fakeVoteRPC fakeVoteRPC
}

type fakeMember struct {
	GetSetLeaderIDFn func(leaderID string)
	GetSelfFn        func() member.Entry
	member.RaftMember
}

func (fakeMember) List() ([]member.Entry, error) {
	panic("implement me")
}

func (fakeMember) Leader() member.Entry {
	panic("implement me")
}

func (fm fakeMember) Self() member.Entry {
	return fm.GetSelfFn()
}

func (fakeMember) SetSelfToLeader() {
	panic("implement me")
}

func (fm fakeMember) SetLeaderID(leaderID string) {
	fm.GetSetLeaderIDFn(leaderID)
}

func (fakeMember) GetLeaderID() (leaderID string) {
	panic("implement me")
}

func (fakeMember) VotedFor() (candidateID string) {
	panic("implement me")
}

func (fakeMember) SetVotedFor(candidateID string) {
	panic("implement me")
}

type fakeRaftState struct {
	GetSetStateFn func(state raftstate.State)
	raftstate.RaftState
}

func (frs fakeRaftState) GetStateChan() <-chan raftstate.State {
	panic("implement me")
}

func (frs fakeRaftState) SetState(state raftstate.State) {
	frs.GetSetStateFn(state)
}

type fakeRaftTerm struct {
	GetIncTermFn func() int64
	raftterm.RaftTerm
}

func (fakeRaftTerm) GetTerm() int64 {
	panic("implement me")
}

func (frt fakeRaftTerm) IncrementTerm() int64 {
	return frt.GetIncTermFn()
}

type fakeRaftVoter struct {
	GetRequestVoteFn func(term int64) <-chan raftvoter.VoteStatus
	raftvoter.RaftVoter
}

func (frv fakeRaftVoter) RequestVote(term int64) <-chan raftvoter.VoteStatus {
	return frv.GetRequestVoteFn(term)
}

type fakeRaftTimer struct {
	GetDeadlineFn    func() time.Time
	GetSetDeadlineFn func(currentTime time.Time)
	rafttimer.RaftTimer
}

func (frt fakeRaftTimer) SetDeadline(currentTime time.Time) {
	frt.GetSetDeadlineFn(currentTime)
}

func (frt fakeRaftTimer) GetDeadline() time.Time {
	return frt.GetDeadlineFn()
}

func (fakeRaftTimer) SetIdleTimeout() {
	panic("implement me")
}

func (fakeRaftTimer) GetIdleTimeout() time.Time {
	panic("implement me")
}

type fakeAeRPC struct {
	GetRaftRpcReqChanFn func() <-chan raftrpc.RaftRpcRequest
	GetProcessFn        func(meta raftrpc.RaftRpcMeta) (raftrpc.RaftRpcResponse, error)
	raftrpc.RaftRpc
}

func (fakeAeRPC) Receive(request raftrpc.RaftRpcRequest) {
	panic("implement me")
}

func (far fakeAeRPC) Process(meta raftrpc.RaftRpcMeta) (raftrpc.RaftRpcResponse, error) {
	return far.GetProcessFn(meta)
}

func (far fakeAeRPC) RaftRpcReqChan() <-chan raftrpc.RaftRpcRequest {
	return far.GetRaftRpcReqChanFn()
}

type fakeVoteRPC struct {
	GetProcessFn        func(meta raftrpc.RaftRpcMeta) (raftrpc.RaftRpcResponse, error)
	GetRaftRpcReqChanFn func() <-chan raftrpc.RaftRpcRequest
	raftrpc.RaftRpc
}

func (fvr fakeVoteRPC) RaftRpcReqChan() <-chan raftrpc.RaftRpcRequest {
	return fvr.GetRaftRpcReqChanFn()
}

func (fvr fakeVoteRPC) Process(meta raftrpc.RaftRpcMeta) (raftrpc.RaftRpcResponse, error) {
	return fvr.GetProcessFn(meta)
}
