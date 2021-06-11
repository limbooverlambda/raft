package log

import (
	"reflect"
	"testing"
)

func TestRaftLog_AllOperations(t *testing.T) {
	entries := []Entry{
		{
			Term:    1,
			Payload: []byte("Hello"),
		},
		{
			Term:    2,
			Payload: []byte("world"),
		},
	}
	l := NewRaftLog("allops_test")
	defer func() {
		deleteFile("allops_test")
	}()
	for _, entry := range entries {
		l.AppendEntry(entry)
	}
	lastLogEntry, err :=  l.LastLogEntryMeta()
	expectErr(t, err, nil)
	equal(t, uint64(2), lastLogEntry.LogIndex)

	indexEntry, err := l.LogEntry(lastLogEntry.LogIndex)
	expectErr(t, err, nil)
	actualPayload := string(indexEntry.Payload)
	equal(t, "world", actualPayload)

	indexEntry, err = l.LogEntry(1)
	expectErr(t, err, nil)
	actualPayload = string(indexEntry.Payload)
	equal(t, "Hello", actualPayload)

	logEntryMeta, err :=  l.LogEntryMeta(1)
	expectErr(t, err, nil)
	equal(t, uint64(1), logEntryMeta.Term)

	err = l.Truncate(2)
	expectErr(t, err, nil)

	logEntryMeta, err =  l.LastLogEntryMeta()
	expectErr(t, err, nil)
	equal(t, uint64(1), logEntryMeta.Term)
}

func TestRaftLog_AppendEntry(t *testing.T) {
	l := NewRaftLog("ae_test")
	defer func() {
		deleteFile("ae_test")
	}()
	entry := Entry{
		Term:    1,
		Payload: []byte("hello"),
	}
	entryMeta, err := l.AppendEntry(entry)
	expectErr(t, nil, err)
	equal(t, EntryMeta{
		Term:         1,
		LogIndex:     1,
		PrevLogTerm:  0,
		PrevLogIndex: 0,
		PayloadSize:  5,
	}, entryMeta)

	nextEntry := Entry{
		Term:    2,
		Payload: []byte("worlds!"),
	}
	nEntryMeta, err := l.AppendEntry(nextEntry)
	expectErr(t, nil, err)
	equal(t, EntryMeta{
		Term:         2,
		LogIndex:     2,
		PrevLogTerm:  1,
		PrevLogIndex: 1,
		PayloadSize:  7,
	}, nEntryMeta)
}

func TestRaftLog_WriteReadIndex(t *testing.T) {
	l := initRaftLog("idx_test")
	defer func() {
		deleteFile("idx_test")
	}()
	index := Index{
		Position:    32,
		Term:        1,
		PayloadSize: 123213,
	}
	i, err := l.writeToIndex(index)
	expectErr(t, err, nil)
	equal(t, i, index)

	_, err = l.readFromIndex(0)
	equal(t, err, ErrIllegalIndex)

	idx, err := l.readFromIndex(1)
	equal(t, idx, index)
	expectErr(t, err, nil)
}

func TestRaftLog_WriteReadLog(t *testing.T) {
	l := initRaftLog("log_test")
	defer func() {
		deleteFile("log_test")
	}()
	size, err := l.writeToLog([]byte("hello"))
	expectErr(t, err, nil)
	equal(t, size, 5)

	currentSize := uint64(size)
	payload, err := l.readFromLog(0, currentSize)
	expectErr(t, err, nil)
	equal(t, string(payload), "hello")

}

func equal(t *testing.T, expected interface{}, actual interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected,actual)  {
		t.Errorf("expected %+v, actual %+v", expected, actual)
	}
}

func expectErr(t *testing.T, err error, expectedError error) {
	t.Helper()
	if err != expectedError {
		t.Errorf("expected error %v, actual error %v", expectedError, err)
	}
}

/*
	//log := NewRaftLog("foo")
	//var index, term uint64 = 0, 0
	//
	//entry := Entry{
	//	Term:    term,
	//	Payload: nil,
	//}
	//_, err := log.AppendEntry(entry)
	//expectErr(t, err, ErrEmptyPayloadAppend)
	//
	////add a payload
	//entry.Payload = []byte("foo")
	//index++
	//expectedResp := EntryMeta{
	//	Term:         entry.Term,
	//	LogIndex:     index,
	//	PrevLogTerm:  entry.Term,
	//	PrevLogIndex: index - 1,
	//}
	//resp, err := log.AppendEntry(entry)
	//expectErr(t, err, nil)
	//equal(t, expectedResp, resp)

	//expectedLogEntry := EntryMeta{
	//	Index:       index,
	//	Term:        0,
	//	PayloadSize: 0,
	//}
	//logEntry := log.GetCurrentLogEntry()
	//equal(t, expectedLogEntry, logEntry)
	//
	//actualIndex := log.GetCurrentLogIndex()
	//equal(t, index, actualIndex)
	//
	//actualEntry, actualErr := log.GetLogEntryAtIndex(index)
	//expectErr(t, actualErr, ErrIllegalIndex)
	//equal(t, actualEntry, Entry{})
	//
	//actualEntry, actualErr = log.GetLogEntryAtIndex(index-1)
	//t.Helper()
	//expectErr(t, actualErr, nil)

*/
