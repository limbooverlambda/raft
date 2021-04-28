package log

type RaftLog interface {
	AppendEntry(entry []byte) (int64, error)
}

func NewRaftLog() RaftLog {
	return raftLog{}
}

type raftLog struct{}

func (raftLog) AppendEntry(entry []byte) (int64, error) {
	panic("implement me")
}
