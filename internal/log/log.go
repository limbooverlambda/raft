package log

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

var (
	byteorder = binary.BigEndian
)

const (
	recordLengthInBytes   = 8
	metadataLengthInBytes = 24
)

type EntryMeta struct {
	Index       uint64
	Term        uint64
	PayloadSize uint64
}

type Entry struct {
	Meta    EntryMeta
	Payload []byte
}

type AppendEntryResponse struct {
	Term         int64
	LogIndex     uint64
	LogSize      uint64
	LogOffset    uint64
	PrevLogTerm  int64
	PrevLogIndex uint64
}

type RaftLog interface {
	AppendEntry(entry Entry) (AppendEntryResponse, error)
	ReadLastEntryMeta() (EntryMeta, error)
}

func NewRaftLog(logID string) RaftLog {
	f, size, err := openLogFile(logID)
	if err != nil {
		panic(fmt.Errorf("unable to open log file due to %v", err))
	}
	return &raftLog{
		logID: logID,
		File:  f,
		buf:   bufio.NewWriter(f),
		size:  uint64(size),
	}
}

func openLogFile(id string) (f *os.File, size int64, err error) {
	f, err = os.OpenFile(id,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, 0, err
	}
	fi, err := os.Stat(id)
	if err != nil {
		return nil, 0, err
	}
	return f, fi.Size(), nil
}

type raftLog struct {
	logID string
	*os.File
	sync.Mutex
	buf      *bufio.Writer
	size     uint64
	position uint64
}

func (rl *raftLog) ReadLastEntryMeta() (EntryMeta, error) {
	rl.Lock()
	defer rl.Unlock()
	metadataBytes := make([]byte, metadataLengthInBytes)

	_, err := rl.File.ReadAt(metadataBytes, int64(rl.position))
	if err != nil {
		return EntryMeta{}, err
	}

	return EntryMeta{
		Index:       byteorder.Uint64(metadataBytes[0:recordLengthInBytes]),
		Term:        byteorder.Uint64(metadataBytes[recordLengthInBytes: recordLengthInBytes+8]),
		PayloadSize: byteorder.Uint64(metadataBytes[recordLengthInBytes+8: recordLengthInBytes+16]),
	}, nil
}

func (rl *raftLog) AppendEntry(entry Entry) (AppendEntryResponse, error) {
	var prevEntryIndex, prevEntryTerm uint64
	prevEntryMeta, err :=  rl.ReadLastEntryMeta()
	if err != nil {
		if err != io.EOF {
			log.Println("Unable to read previous value")
			return AppendEntryResponse{}, err
		}
	}
	prevEntryTerm = prevEntryMeta.Term
	prevEntryIndex = prevEntryMeta.Index

	rl.Lock()
	defer rl.Unlock()


	rl.position = rl.size

	entry.Meta.Index = prevEntryIndex + 1
	if err := binary.Write(rl.buf, byteorder, entry.Meta.Index); err != nil {
		return AppendEntryResponse{}, err
	}

	if err := binary.Write(rl.buf, byteorder, entry.Meta.Term); err != nil {
		return AppendEntryResponse{}, err
	}

	if err := binary.Write(rl.buf, byteorder, entry.Meta.PayloadSize); err != nil {
		return AppendEntryResponse{}, err
	}

	w, err := rl.buf.Write(entry.Payload)
	if err != nil {
		return AppendEntryResponse{}, err
	}
	//Flush to the disc right away
	err = rl.buf.Flush()
	if err != nil {
		return AppendEntryResponse{}, err
	}

	w += metadataLengthInBytes
	rl.size += uint64(w)

	return AppendEntryResponse{
		Term:         int64(entry.Meta.Term),
		LogSize:      rl.size,
		LogOffset:    entry.Meta.Index,
		PrevLogTerm:  int64(prevEntryTerm),
		PrevLogIndex: prevEntryIndex,
	}, nil
}
