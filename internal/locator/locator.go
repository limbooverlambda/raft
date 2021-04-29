package locator

import (
	appendentrysender "github.com/kitengo/raft/internal/appendentry"
	raftapplicator "github.com/kitengo/raft/internal/applicator"
	raftheartbeat "github.com/kitengo/raft/internal/heartbeat"
	raftlog "github.com/kitengo/raft/internal/log"
	raftmember "github.com/kitengo/raft/internal/member"
	raftrpc "github.com/kitengo/raft/internal/rpc"
	raftstate "github.com/kitengo/raft/internal/state"
	raftterm "github.com/kitengo/raft/internal/term"
	rafttimer "github.com/kitengo/raft/internal/timer"
	raftvoter "github.com/kitengo/raft/internal/voter"
)

type ServiceLocator interface {
	GetRpcLocator() RpcLocator
	GetRaftState() raftstate.RaftState
	GetRaftTerm() raftterm.RaftTerm
	GetRaftTimer() rafttimer.RaftTimer
	GetRaftVoter() raftvoter.RaftVoter
	GetRaftHeartbeat() raftheartbeat.RaftHeartbeat
}

func NewServiceLocator() ServiceLocator {
	term := raftterm.NewRaftTerm()
	return &serviceLocator{
		raftState:     raftstate.NewRaftState(),
		raftTerm:      term,
		raftVoter:     raftvoter.NewRaftVoter(),
		raftHeartbeat: raftheartbeat.NewRaftHeartbeat(),
		rpcLocator:    NewRpcLocator(term),
	}
}

type serviceLocator struct {
	raftState     raftstate.RaftState
	raftTerm      raftterm.RaftTerm
	raftVoter     raftvoter.RaftVoter
	raftHeartbeat raftheartbeat.RaftHeartbeat
	rpcLocator    RpcLocator
}

func (sl *serviceLocator) GetRaftVoter() raftvoter.RaftVoter {
	return sl.raftVoter
}

func (sl *serviceLocator) GetRaftTimer() rafttimer.RaftTimer {
	return rafttimer.NewRaftTimer()
}

func (sl *serviceLocator) GetRaftTerm() raftterm.RaftTerm {
	return sl.raftTerm
}

func (sl *serviceLocator) GetRaftState() raftstate.RaftState {
	return sl.raftState
}

func (sl *serviceLocator) GetRpcLocator() RpcLocator {
	return sl.rpcLocator
}

func (sl *serviceLocator) GetRaftHeartbeat() raftheartbeat.RaftHeartbeat {
	return sl.raftHeartbeat
}

type RpcLocator interface {
	GetAppendEntrySvc() raftrpc.RaftRpc
	GetRequestVoteSvc() raftrpc.RaftRpc
	GetClientCommandSvc() raftrpc.RaftRpc
}

func NewRpcLocator(term raftterm.RaftTerm) RpcLocator {
	raftLog := raftlog.NewRaftLog("log")
	raftMember := raftmember.NewRaftMember()
	sender := appendentrysender.NewSender()
	index := raftstate.NewRaftIndex()
	raftApplicator := raftapplicator.NewRaftApplicator()
	return &rpcLocator{
		raftAppend:      raftrpc.NewRaftAppendEntry(),
		raftRequestVote: raftrpc.NewRaftRequestVote(),
		raftClientCommand: raftrpc.NewRaftClientCommand(term,
			raftLog,
			raftMember,
			sender,
			index,
			raftApplicator),
	}
}

type rpcLocator struct {
	raftAppend        raftrpc.RaftRpc
	raftRequestVote   raftrpc.RaftRpc
	raftClientCommand raftrpc.RaftRpc
}

func (rpcl *rpcLocator) GetAppendEntrySvc() raftrpc.RaftRpc {
	return rpcl.raftAppend
}

func (rpcl *rpcLocator) GetRequestVoteSvc() raftrpc.RaftRpc {
	return rpcl.raftRequestVote
}

func (rpcl *rpcLocator) GetClientCommandSvc() raftrpc.RaftRpc {
	return rpcl.raftClientCommand
}
