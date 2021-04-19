package log

type RaftLog interface {
	AppendEntry(entry []byte)
}
