package appendentry

import (
	raftlog "github.com/kitengo/raft/internal/log"
	"github.com/kitengo/raft/internal/models"
)

type fakeLog struct{
	GetLogEntryAtIndexFn func(index uint64) (raftlog.Entry, error)
}

func (fakeLog) AppendEntry(entry raftlog.Entry) (raftlog.EntryMeta, error) {
	panic("implement me")
}

func (fakeLog) LastLogEntryMeta() (raftlog.EntryMeta, error) {
	panic("implement me")
}

func (fl fakeLog) LogEntry(indexID uint64) (raftlog.Entry, error) {
	return fl.GetLogEntryAtIndexFn(indexID)
}

func (fakeLog) LogEntryMeta(indexID uint64) (raftlog.EntryMeta, error) {
	panic("implement me")
}

func (fl fakeLog) Truncate(indexID uint64) error {
	panic("implement me")
}

//func (fl fakeLog) GetCurrentLogIndex() uint64 {
//	panic("implement me")
//}
//
//
//func (fakeLog) GetLogEntryMetaAtIndex(index uint64) (raftlog.EntryMeta, error) {
//	panic("implement me")
//}
//
//func (fl fakeLog) GetLogEntryAtIndex(index uint64) (raftlog.Entry, error) {
//	return fl.GetLogEntryAtIndexFn(index)
//}
//
//func (fakeLog) TruncateFromIndex(index uint64) error {
//	panic("implement me")
//}
//
//func (fakeLog) GetCurrentLogEntry() raftlog.EntryMeta {
//	panic("implement me")
//}

type fakeSender struct {
	GetSendCommandFn func(requestConv models.RequestConverter, ip, port string) (response models.Response, err error)
}

func (f fakeSender) SendCommand(requestConv models.RequestConverter, ip, port string) (response models.Response, err error) {
	return f.GetSendCommandFn(requestConv, ip, port)
}
