package heartbeat

type RaftHeartbeat interface {
	SendHeartbeats()
}

func NewRaftHeartbeat() RaftHeartbeat {
	return raftHeartbeat{}
}

type raftHeartbeat struct{}

func (raftHeartbeat) SendHeartbeats() {
	panic("implement me")
}
