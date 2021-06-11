package heartbeat

import (
	"github.com/kitengo/raft/internal/member"
	"github.com/kitengo/raft/internal/models"
	raftterm "github.com/kitengo/raft/internal/term"
)

type fakeRaftTerm struct {
	GetTermFn    func() uint64
	GetIncTermFn func() uint64
	raftterm.RaftTerm
}

func (frt fakeRaftTerm) GetTerm() uint64 {
	return frt.GetTermFn()
}

func (frt fakeRaftTerm) IncrementTerm() uint64 {
	return frt.GetIncTermFn()
}

type fakeRaftMember struct {
	GetSetLeaderIDFn func(leaderID string)
	GetSelfFn        func() member.Entry
	GetLeaderFn      func() member.Entry
	GetListFn        func() ([]member.Entry, error)
	member.RaftMember
}

func (fm fakeRaftMember) List() ([]member.Entry, error) {
	return fm.GetListFn()
}

func (fm fakeRaftMember) Leader() member.Entry {
	return fm.GetLeaderFn()
}

func (fm fakeRaftMember) Self() member.Entry {
	return fm.GetSelfFn()
}

func (fakeRaftMember) SetSelfToLeader() {
	panic("implement me")
}

func (fm fakeRaftMember) SetLeaderID(leaderID string) {
	fm.GetSetLeaderIDFn(leaderID)
}

func (fakeRaftMember) GetLeaderID() (leaderID string) {
	panic("implement me")
}

func (fakeRaftMember) VotedFor() (candidateID string) {
	panic("implement me")
}

func (fakeRaftMember) SetVotedFor(candidateID string) {
	panic("implement me")
}

type fakeRaftSender struct {
	GetSendCommandFn func(requestConv models.RequestConverter, ip, port string) (response models.Response, err error)
}

func (f fakeRaftSender) SendCommand(requestConv models.RequestConverter, ip, port string) (response models.Response, err error) {
	return f.GetSendCommandFn(requestConv, ip, port)
}
