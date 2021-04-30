package log

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	byteorder = binary.BigEndian
)

const (
	recordLengthInBytes = 8
)

type Entry struct {
	Term    int64
	Payload []byte
}

type AppendEntryResponse struct {
	Term int64
	LogSize uint64
	LogOffset uint64
	PrevLogTerm int64
	PrevLogOffset uint64
}

type RaftLog interface {
	AppendEntry(entry Entry) (AppendEntryResponse, error)
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
	buf  *bufio.Writer
	size uint64
}

func (rl *raftLog) AppendEntry(entry Entry) (AppendEntryResponse, error) {
	log.Println("Appending Entry to log")
	rl.Lock()
	defer rl.Unlock()

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(entry)

	if err != nil {
		return AppendEntryResponse{}, err
	}

	position := rl.size
	payload := buffer.Bytes()

	if err := binary.Write(rl.buf, byteorder, uint64(len(payload))); err != nil {
		return AppendEntryResponse{}, err
	}

	w, err := rl.buf.Write(payload)
	if err != nil {
		return AppendEntryResponse{}, err
	}
	//Flush to the disc right away
	err = rl.buf.Flush()
	if err != nil {
		return AppendEntryResponse{}, err
	}

	w += recordLengthInBytes
	rl.size += uint64(w)

	return AppendEntryResponse{
		Term:          entry.Term,
		LogSize:       rl.size,
		LogOffset:     position,
		PrevLogTerm:   0,
		PrevLogOffset: 0,
	}, nil
}
