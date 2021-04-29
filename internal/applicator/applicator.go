package applicator

import (
	"log"
)

type RaftApplicator interface {
	Apply(state []byte) error
}

func NewRaftApplicator() RaftApplicator {
	return &raftApplicator{}
}

type raftApplicator struct{}

func (*raftApplicator) Apply(state []byte) error {
	log.Println("Applying the append to the state machine")
	return nil
}

