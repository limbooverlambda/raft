package log

type RaftLog interface {
	AppendEntry(entry []byte)
}

func NewRaftLog() RaftLog {
	return raftLog{}
}

type raftLog struct{}

func (raftLog) AppendEntry(entry []byte) {
	panic("implement me")
}
