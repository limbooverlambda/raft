package heartbeat

import (
	"log"
)

type RaftHeartbeat interface {
	SendHeartbeats()
}

func NewRaftHeartbeat() RaftHeartbeat {
	return raftHeartbeat{}
}

type raftHeartbeat struct{}

func (raftHeartbeat) SendHeartbeats() {
	log.Println("Sending heartbeat from leader...")
}
