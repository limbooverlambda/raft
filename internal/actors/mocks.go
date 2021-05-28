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

type followerStub struct {
	fakeState   fakeRaftState
	fakeAeRPC   fakeAeRPC
	fakeVoteRPC fakeVoteRPC
	fakeTimer   fakeRaftTimer
}

type leaderStub struct {
	state     fakeRaftState
	heartbeat fakeRaftHeartbeat
	clientRPC fakeClientRPC
	aeRPC     fakeAeRPC
	voteRPC   fakeVoteRPC
	raftTerm  fakeRaftTerm
	raftTimer fakeRaftTimer
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
	GetTermFn    func() int64
	GetIncTermFn func() int64
	raftterm.RaftTerm
}

func (frt fakeRaftTerm) GetTerm() int64 {
	return frt.GetTermFn()
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
	GetSetIdleTimeFn func()
	GetIdleTimoutFn  func() time.Time
	rafttimer.RaftTimer
}

func (frt fakeRaftTimer) SetDeadline(currentTime time.Time) {
	frt.GetSetDeadlineFn(currentTime)
}

func (frt fakeRaftTimer) GetDeadline() time.Time {
	return frt.GetDeadlineFn()
}

func (frt fakeRaftTimer) SetIdleTimeout() {
	frt.GetSetIdleTimeFn()
}

func (frt fakeRaftTimer) GetIdleTimeout() time.Time {
	return frt.GetIdleTimoutFn()
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

type fakeClientRPC struct {
	GetProcessFn        func(meta raftrpc.RaftRpcMeta) (raftrpc.RaftRpcResponse, error)
	GetRaftRpcReqChanFn func() <-chan raftrpc.RaftRpcRequest
}

func (fakeClientRPC) Receive(request raftrpc.RaftRpcRequest) {
	panic("implement me")
}

func (fcr fakeClientRPC) Process(meta raftrpc.RaftRpcMeta) (raftrpc.RaftRpcResponse, error) {
	return fcr.GetProcessFn(meta)
}

func (fcr fakeClientRPC) RaftRpcReqChan() <-chan raftrpc.RaftRpcRequest {
	return fcr.GetRaftRpcReqChanFn()
}

type fakeRaftHeartbeat struct {
	GetSendHeartBeatFn func()
}

func (frh fakeRaftHeartbeat) SendHeartbeats() {
	frh.GetSendHeartBeatFn()
}

type candidateHelperStub struct {
	voteStatusChan   chan raftvoter.VoteStatus
	aeRequestChan    chan raftrpc.RaftRpcRequest
	voteRequestsChan chan raftrpc.RaftRpcRequest
	candidateStub
}

type followerHelperStub struct {
	aeRequestChan    chan raftrpc.RaftRpcRequest
	voteRequestsChan chan raftrpc.RaftRpcRequest
	followerStub
}

type leaderHelperStub struct {
	aeRequestChan     chan raftrpc.RaftRpcRequest
	voteRequestsChan  chan raftrpc.RaftRpcRequest
	clientRequestChan chan raftrpc.RaftRpcRequest
	leaderStub
}

func createFollowerStub() followerHelperStub {
	aeRequests := make(chan raftrpc.RaftRpcRequest, 1)
	voteRequests := make(chan raftrpc.RaftRpcRequest, 1)
	fStub := followerStub{
		fakeState:   getFakeRaftState(),
		fakeAeRPC:   getFakeAeRPC(aeRequests),
		fakeVoteRPC: getFakeVoteRPC(voteRequests),
		fakeTimer:   getFakeRaftTimer(),
	}
	return followerHelperStub{
		aeRequestChan:    aeRequests,
		voteRequestsChan: voteRequests,
		followerStub:     fStub,
	}
}

func createCandidateStub() candidateHelperStub {
	voteStatuses := make(chan raftvoter.VoteStatus, 1)
	aeRequests := make(chan raftrpc.RaftRpcRequest, 1)
	voteRequests := make(chan raftrpc.RaftRpcRequest, 1)

	cStub := candidateStub{
		fakeMember:  getFakeMember(),
		fakeState:   getFakeRaftState(),
		fakeTerm:    getFakeRaftTerm(),
		fakeVoter:   getFakeRaftVoter(voteStatuses),
		fakeTimer:   getFakeRaftTimer(),
		fakeAeRPC:   getFakeAeRPC(aeRequests),
		fakeVoteRPC: getFakeVoteRPC(voteRequests),
	}
	return candidateHelperStub{
		voteStatusChan:   voteStatuses,
		aeRequestChan:    aeRequests,
		voteRequestsChan: voteRequests,
		candidateStub:    cStub,
	}
}

func createLeaderStub() leaderHelperStub {
	aeRequests := make(chan raftrpc.RaftRpcRequest, 1)
	voteRequests := make(chan raftrpc.RaftRpcRequest, 1)
	clientRequests := make(chan raftrpc.RaftRpcRequest, 1)
	lStub := leaderStub{
		state:     getFakeRaftState(),
		heartbeat: getFakeRaftHeartbeat(),
		clientRPC: getFakeClientRPC(clientRequests),
		aeRPC:     getFakeAeRPC(aeRequests),
		voteRPC:   getFakeVoteRPC(voteRequests),
		raftTerm:  getFakeRaftTerm(),
		raftTimer: getFakeRaftTimer(),
	}
	return leaderHelperStub{
		aeRequestChan:     aeRequests,
		voteRequestsChan:  voteRequests,
		clientRequestChan: clientRequests,
		leaderStub:        lStub,
	}
}

func getFakeVoteRPC(voteRequests chan raftrpc.RaftRpcRequest) fakeVoteRPC {
	return fakeVoteRPC{
		GetRaftRpcReqChanFn: func() <-chan raftrpc.RaftRpcRequest {
			return voteRequests
		},
	}
}

func getFakeAeRPC(aeRequests chan raftrpc.RaftRpcRequest) fakeAeRPC {
	return fakeAeRPC{
		GetRaftRpcReqChanFn: func() <-chan raftrpc.RaftRpcRequest {
			return aeRequests
		},
	}
}

func getFakeClientRPC(clientRequests chan raftrpc.RaftRpcRequest) fakeClientRPC {
	return fakeClientRPC{
		GetRaftRpcReqChanFn: func() <-chan raftrpc.RaftRpcRequest {
			return clientRequests
		},
	}
}

func getFakeRaftTimer() fakeRaftTimer {
	return fakeRaftTimer{
		GetDeadlineFn: func() time.Time {
			return time.Now()
		},
		GetSetDeadlineFn: func(currentTime time.Time) {
			//NO-OP
		},
	}
}

func getFakeRaftVoter(voteStatuses chan raftvoter.VoteStatus) fakeRaftVoter {
	return fakeRaftVoter{
		GetRequestVoteFn: func(term int64) <-chan raftvoter.VoteStatus {
			return voteStatuses
		},
	}
}

func getFakeRaftTerm() fakeRaftTerm {
	return fakeRaftTerm{
		GetIncTermFn: func() int64 {
			return 1
		},
	}
}

func getFakeRaftState() fakeRaftState {
	return fakeRaftState{
		GetSetStateFn: func(state raftstate.State) {
		},
	}
}

func getFakeMember() fakeMember {
	return fakeMember{
		GetSetLeaderIDFn: func(leaderID string) {

		},
		GetSelfFn: func() member.Entry {
			return member.Entry{}
		},
	}
}

func getFakeRaftHeartbeat() fakeRaftHeartbeat {
	return fakeRaftHeartbeat{}
}
