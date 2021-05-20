package member

import (
	"errors"
	"github.com/kitengo/raft/internal/rconfig"
	"log"
	"sync"
)

type EntryType int

const (
	Self EntryType = iota
	Leader
	Members
	VotedFor
)

type Entry struct {
	ID      string
	Address string
	Port    string
}

type RaftMember interface {
	List() ([]Entry, error)
	Leader() Entry
	Self() Entry
	SetSelfToLeader()
	SetLeaderID(leaderID string)
	GetLeaderID() (leaderID string)
	VotedFor() (candidateID string)
	SetVotedFor(candidateID string)
}

func NewRaftMember(config rconfig.Config) RaftMember {
	memberMap := sync.Map{}
	self := Entry{
		ID:      config.ServerID,
		Address: config.ServerHost,
		Port:    config.ServerPort,
	}
	var memberEntries []Entry
	for _, mc := range config.MemberConfig {
		entry := Entry{
			ID:      mc.ID,
			Address: mc.IP,
			Port:    mc.Port,
		}
		log.Printf("member entry %+v\n", entry)
		memberEntries = append(memberEntries, entry)
	}
	memberMap.Store(Self, self)
	memberMap.Store(Members, memberEntries)
	memberMap.Store(VotedFor, "")
	return &raftMember{
		Map: &memberMap,
	}
}

type raftMember struct {
	*sync.Map
	leaderID string
}

func (rm *raftMember) GetLeaderID() (leaderID string) {
	leaderID = rm.leaderID
	return
}

func (rm *raftMember) SetLeaderID(leaderID string) {
	rm.leaderID = leaderID
}

func (rm *raftMember) VotedFor() (candidateID string) {
	if v, ok := rm.Load(VotedFor); ok {
		candidateID = v.(string)
	}
	return
}

func (rm *raftMember) SetVotedFor(candidateID string) {
	rm.Store(VotedFor, candidateID)
}

func (rm *raftMember) SetSelfToLeader() {
	rm.Store(Leader, rm.Self())
}

func (rm *raftMember) Self() Entry {
	s, ok := rm.Load(Self)
	if !ok {
		panic("failed to find self")
	}
	return s.(Entry)
}

func (rm *raftMember) Leader() Entry {
	s, ok := rm.Load(Leader)
	if !ok {
		panic("failed to find leader")
	}
	return s.(Entry)
}

func (rm *raftMember) List() ([]Entry, error) {
	if s, ok := rm.Load(Members); ok {
		return s.([]Entry), nil
	}
	return nil, errors.New("unable to find members")
}
