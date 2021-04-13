package peers

type RaftPeers interface {
	SendHeartbeat()
}
