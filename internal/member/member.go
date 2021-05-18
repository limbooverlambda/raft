package member

import (
	"errors"
	"github.com/kitengo/raft/internal/rconfig"
	"sync"
)


type EntryType int

const (
	Self EntryType = iota
	Leader
	Members
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
		memberEntries = append(memberEntries, entry)
	}
	memberMap.Store(Self, self)
	memberMap.Store(Members, memberEntries)
	return &raftMember{
		memberMap,
	}
}

type raftMember struct{
	sync.Map
}

func (rm *raftMember) Self() Entry {
	s, ok :=  rm.Load(Self)
	if !ok {
		panic("failed to find self")
	}
	return s.(Entry)
}

func (rm *raftMember) Leader() Entry {
	s, ok :=  rm.Load(Leader)
	if !ok {
		panic("failed to find self")
	}
	return s.(Entry)
}

func (rm *raftMember) List() ([]Entry, error) {
	if s, ok := rm.Load(Members); ok {
		return s.([]Entry), nil
	}
	return nil, errors.New("unable to find members")
}
