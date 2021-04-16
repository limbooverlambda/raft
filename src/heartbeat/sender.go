package heartbeat

type RaftHeartbeat interface {
	SendHeartbeats()
}
