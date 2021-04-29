package member

import (
	"log"
)

type Entry struct {
	ID string
}

type RaftMember interface {
	List() ([]Entry, error)
}

func NewRaftMember() RaftMember {
	return &raftMember{}
}

type raftMember struct{}

func (rm *raftMember) List() ([]Entry, error) {
	log.Println("Listing the members of the cluster")
	return []Entry{{"id1"},{"id2"}, {"id3"}}, nil
}
