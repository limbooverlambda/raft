package log

import (
	"testing"
)

func Test_raftLog_AppendEntry(t *testing.T) {
	log := NewRaftLog("test")
	resp, err := log.AppendEntry(Entry{
		Meta:    EntryMeta{
			Index:       0,
			Term:        1,
			PayloadSize: 0,
		},
		Payload: []byte{},
	})
	if err != nil {
		t.Log("Got err", err)
	}
	t.Log("AE response1", resp)
	entryMeta, err := log.ReadLastEntryMeta()
	if err != nil {
		t.Log("Got err", err)
	}
	t.Log("Entry meta is 1", entryMeta)
	resp1, err := log.AppendEntry(Entry{
		Meta:    EntryMeta{
			Index:       1,
			Term:        1,
			PayloadSize: 0,
		},
		Payload: []byte{},
	})
	if err != nil {
		t.Log("Got err", err)
	}
	t.Log("AE response 2", resp1)
	entryMeta1, err := log.ReadLastEntryMeta()
	if err != nil {
		t.Log("Got err", err)
	}
	t.Log("Entry meta is 2 ", entryMeta1)
}
