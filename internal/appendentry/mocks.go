package appendentry

import (
	raftlog "github.com/kitengo/raft/internal/log"
	"github.com/kitengo/raft/internal/models"
)

type fakeLog struct{}

func (fakeLog) AppendEntry(entry raftlog.Entry) (raftlog.AppendEntryResponse, error) {
	panic("implement me")
}

func (fakeLog) GetLogEntryMetaAtIndex(index uint64) (raftlog.EntryMeta, error) {
	panic("implement me")
}

func (fakeLog) GetLogEntryAtIndex(index uint64) (raftlog.Entry, error) {
	panic("implement me")
}

func (fakeLog) TruncateFromIndex(index uint64) error {
	panic("implement me")
}

func (fakeLog) GetCurrentLogEntry() raftlog.EntryMeta {
	panic("implement me")
}

type fakeSender struct {
	GetSendCommandFn func(requestConv models.RequestConverter, ip, port string) (response models.Response, err error)
}

func (f fakeSender) SendCommand(requestConv models.RequestConverter, ip, port string) (response models.Response, err error) {
	return f.GetSendCommandFn(requestConv, ip, port)
}
