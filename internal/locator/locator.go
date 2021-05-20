package locator

import (
	"fmt"
	appendentrysender "github.com/kitengo/raft/internal/appendentry"
	raftapplicator "github.com/kitengo/raft/internal/applicator"
	raftheartbeat "github.com/kitengo/raft/internal/heartbeat"
	raftlog "github.com/kitengo/raft/internal/log"
	raftmember "github.com/kitengo/raft/internal/member"
	"github.com/kitengo/raft/internal/rconfig"
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
	GetRaftMember() raftmember.RaftMember
	GetRaftHeartbeat() raftheartbeat.RaftHeartbeat
}

func NewServiceLocator(config rconfig.Config) ServiceLocator {
	raftTerm := raftterm.NewRaftTerm()
	raftMember := raftmember.NewRaftMember(config)
	index := raftstate.NewRaftIndex()
	raftLog := raftlog.NewRaftLog(fmt.Sprintf("log-%s", config.ServerID))
	return &serviceLocator{
		raftState:     raftstate.NewRaftState(),
		raftTerm:      raftTerm,
		raftVoter:     raftvoter.NewRaftVoter(raftMember, raftLog, raftTerm),
		raftHeartbeat: raftheartbeat.NewRaftHeartbeat(raftTerm, raftMember, index),
		raftMember:    raftMember,
		rpcLocator:    NewRpcLocator(raftTerm, raftMember, index, raftLog),
	}
}

type serviceLocator struct {
	raftState     raftstate.RaftState
	raftTerm      raftterm.RaftTerm
	raftVoter     raftvoter.RaftVoter
	raftHeartbeat raftheartbeat.RaftHeartbeat
	raftMember    raftmember.RaftMember
	rpcLocator    RpcLocator
}

func (sl *serviceLocator) GetRaftMember() raftmember.RaftMember {
	return sl.raftMember
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

func NewRpcLocator(term raftterm.RaftTerm,
	raftMember raftmember.RaftMember,
	raftIndex *raftstate.RaftIndex,
	raftLog raftlog.RaftLog) RpcLocator {

	sender := appendentrysender.NewSender(raftLog)

	raftApplicator := raftapplicator.NewRaftApplicator()
	return &rpcLocator{
		raftAppend:      raftrpc.NewRaftAppendEntry(term, raftLog, raftIndex),
		raftRequestVote: raftrpc.NewRaftRequestVote(term, raftLog, raftMember),
		raftClientCommand: raftrpc.NewRaftClientCommand(term,
			raftLog,
			raftMember,
			sender,
			raftIndex,
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
