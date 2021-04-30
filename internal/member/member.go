package member

import (
	"log"
)

type Entry struct {
	ID string
	Address string
}

type RaftMember interface {
	List() ([]Entry, error)
	Leader() Entry
}

func NewRaftMember() RaftMember {
	return &raftMember{}
}

type raftMember struct{}

func (rm *raftMember) Leader() Entry {
	return Entry{
		ID:      "id4",
		Address: "",
	}
}

func (rm *raftMember) List() ([]Entry, error) {
	log.Println("Listing the members of the cluster")
	return []Entry{{ID:"id1"},{ID:"id2"}, {ID:"id3"}}, nil
}
