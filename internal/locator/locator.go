package locator

import (
	raftheartbeat "github.com/kitengo/raft/internal/heartbeat"
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
	return &serviceLocator{
		raftState: raftstate.NewRaftState(),
		raftTerm: raftterm.NewRaftTerm(),
		raftVoter: raftvoter.NewRaftVoter(),
		raftHeartbeat: raftheartbeat.NewRaftHeartbeat(),
	}
}

type serviceLocator struct{
	raftState raftstate.RaftState
	raftTerm raftterm.RaftTerm
	raftVoter raftvoter.RaftVoter
	raftHeartbeat raftheartbeat.RaftHeartbeat
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

func (*serviceLocator) GetRpcLocator() RpcLocator {
	return rpcLocator{}
}

func (sl *serviceLocator) GetRaftHeartbeat() raftheartbeat.RaftHeartbeat {
	return sl.raftHeartbeat
}

type RpcLocator interface {
	GetAppendEntrySvc() raftrpc.RaftAppendEntry
	GetRequestVoteSvc() raftrpc.RaftRequestVote
	GetClientCommandSvc() raftrpc.RaftClientCommand
}

type rpcLocator struct{}

func (rpcLocator) GetAppendEntrySvc() raftrpc.RaftAppendEntry {
	return raftrpc.NewRaftAppendEntry()
}

func (rpcLocator) GetRequestVoteSvc() raftrpc.RaftRequestVote {
	return raftrpc.NewRaftRequestVote()
}

func (rpcLocator) GetClientCommandSvc() raftrpc.RaftClientCommand {
	return raftrpc.NewRaftClientCommand()
}
