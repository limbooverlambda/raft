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
		rpcLocator: NewRpcLocator(),
	}
}

type serviceLocator struct{
	raftState raftstate.RaftState
	raftTerm raftterm.RaftTerm
	raftVoter raftvoter.RaftVoter
	raftHeartbeat raftheartbeat.RaftHeartbeat
	rpcLocator RpcLocator
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
	GetAppendEntrySvc() raftrpc.RaftAppendEntry
	GetRequestVoteSvc() raftrpc.RaftRequestVote
	GetClientCommandSvc() raftrpc.RaftClientCommand
}

func NewRpcLocator() RpcLocator {
	return &rpcLocator{
		raftAppend:        raftrpc.NewRaftAppendEntry(),
		raftRequestVote:   raftrpc.NewRaftRequestVote(),
		raftClientCommand: raftrpc.NewRaftClientCommand(),
	}
}

type rpcLocator struct{
	raftAppend raftrpc.RaftAppendEntry
	raftRequestVote raftrpc.RaftRequestVote
	raftClientCommand raftrpc.RaftClientCommand

}

func (rpcl *rpcLocator) GetAppendEntrySvc() raftrpc.RaftAppendEntry {
	return rpcl.raftAppend
}

func (rpcl *rpcLocator) GetRequestVoteSvc() raftrpc.RaftRequestVote {
	return rpcl.raftRequestVote
}

func (rpcl *rpcLocator) GetClientCommandSvc() raftrpc.RaftClientCommand {
	return rpcl.raftClientCommand
}
