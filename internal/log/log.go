package log

import (
	"bufio"
	"encoding/binary"
	"fmt"
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
	Index   uint64
	Term    uint64
	Payload []byte
}

type AppendEntryResponse struct {
	Term         int64
	LogIndex     uint64
	PrevLogTerm  int64
	PrevLogIndex uint64
}

//RaftLog is an abstraction around the index and the raw commands to the
//replicated state machine.
// - Create an index file
// - Create the raw log file
// - Add the EntryMeta to the index file
type RaftLog interface {
	AppendEntry(entry Entry) (AppendEntryResponse, error)
	GetLogEntryMetaAtIndex(index uint64) (EntryMeta, error)
	GetLogEntryAtIndex(index uint64) (Entry, error)
	TruncateFromIndex(index uint64) error
	GetCurrentLogEntry() EntryMeta
	GetCurrentLogIndex() uint64
}

func NewRaftLog(logID string) RaftLog {
	logFile, logSize, err := openFile(logID)
	if err != nil {
		panic(fmt.Errorf("unable to open log file due to %v", err))
	}
	idxFile, idxSize, err := openFile(logID + "_idx")
	if err != nil {
		panic(fmt.Errorf("unable to open index file due to %v", err))
	}
	return &raftLog{
		logID:   logID,
		logFile: logFile,
		logBuf:  bufio.NewWriter(logFile),
		logSize: uint64(logSize),
		idxFile: idxFile,
		idxBuf:  bufio.NewWriter(idxFile),
		idxSize: uint64(idxSize),
	}
}

func openFile(id string) (f *os.File, size int64, err error) {
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
	sync.RWMutex
	logFile     *os.File
	logBuf      *bufio.Writer
	idxFile     *os.File
	idxBuf      *bufio.Writer
	idxSize     uint64
	logSize     uint64
	logPosition uint64
	logTerm     uint64
	logIndex    uint64
}

func (rl *raftLog) GetCurrentLogIndex() uint64 {
	return rl.logIndex
}

func (rl *raftLog) GetCurrentLogEntry() EntryMeta {
	return EntryMeta{
		Index: rl.logIndex,
		Term:  rl.logTerm,
	}
}

func (rl *raftLog) GetLogEntryMetaAtIndex(index uint64) (EntryMeta, error) {
	rl.RLock()
	defer rl.RUnlock()
	metadataBytes := make([]byte, metadataLengthInBytes)
	offset := (index - 1) * metadataLengthInBytes
	_, err := rl.idxFile.ReadAt(metadataBytes, int64(offset))
	if err != nil {
		return EntryMeta{}, err
	}
	_, logTerm, _ := byteorder.Uint64(metadataBytes[0:recordLengthInBytes]),
		byteorder.Uint64(metadataBytes[recordLengthInBytes:recordLengthInBytes+8]),
		byteorder.Uint64(metadataBytes[recordLengthInBytes+8:recordLengthInBytes+16])

	return EntryMeta{
		Index: index,
		Term:  logTerm,
		//PayloadSize: logPayloadSize,
	}, nil
}

func (rl *raftLog) GetLogEntryAtIndex(index uint64) (Entry, error) {
	rl.RLock()
	defer rl.RUnlock()
	metadataBytes := make([]byte, metadataLengthInBytes)
	offset := (index - 1) * metadataLengthInBytes
	_, err := rl.idxFile.ReadAt(metadataBytes, int64(offset))
	if err != nil {
		return Entry{}, err
	}
	logPosition, logTerm, logEntrySize := byteorder.Uint64(metadataBytes[0:recordLengthInBytes]),
		byteorder.Uint64(metadataBytes[recordLengthInBytes:recordLengthInBytes+8]),
		byteorder.Uint64(metadataBytes[recordLengthInBytes+8:recordLengthInBytes+16])
	payloadBytes := make([]byte, logEntrySize)
	_, err = rl.logFile.ReadAt(payloadBytes, int64(logPosition))
	if err != nil {
		return Entry{}, err
	}
	return Entry{
		Index:   index,
		Term:    logTerm,
		Payload: payloadBytes,
	}, nil
}

func (rl *raftLog) TruncateFromIndex(index uint64) error {
	rl.Lock()
	defer rl.Unlock()
	idxTruncationSize := rl.idxSize - (index-1)*metadataLengthInBytes
	if err := rl.idxFile.Truncate(int64(idxTruncationSize)); err != nil {
		return err
	}
	rl.idxSize = idxTruncationSize
	rl.logIndex = index - 1
	return nil
}

func (rl *raftLog) AppendEntry(entry Entry) (AppendEntryResponse, error) {
	rl.Lock()
	defer rl.Unlock()

	logPosition := rl.logSize
	w, err := rl.logBuf.Write(entry.Payload)
	if err != nil {
		log.Printf("Failed to write payload %v\n", err)
		return AppendEntryResponse{}, err
	}
	//Flush to the disc right away
	err = rl.logBuf.Flush()
	if err != nil {
		log.Printf("Failed to flush logbuffer payload to disc %v\n", err)
		return AppendEntryResponse{}, err
	}
	rl.logSize += uint64(w)
	//Add the position, term and payload size to the index
	if err := binary.Write(rl.idxBuf, byteorder, logPosition); err != nil {
		log.Printf("Failed to write log position to index %v\n", err)
		return AppendEntryResponse{}, err
	}

	if err := binary.Write(rl.idxBuf, byteorder, entry.Term); err != nil {
		log.Printf("Failed to write entry term to index %v\n", err)
		return AppendEntryResponse{}, err
	}
	//Need to cast to int32 since binary.Write will fail on data that doesn't have a fixed size
	payloadSize := int32(len(entry.Payload))
	if err := binary.Write(rl.idxBuf, byteorder, payloadSize); err != nil {
		log.Printf("Failed to write payload size to index %v\n", err)
		return AppendEntryResponse{}, err
	}

	//Flush to the disc right away
	err = rl.idxBuf.Flush()
	if err != nil {
		log.Printf("Failed to flush indexbuffer payload to disc %v\n", err)
		return AppendEntryResponse{}, err
	}

	//prevLogTerm and prevLogIndex
	prevLogTerm := rl.logTerm
	prevLogIdx := rl.logIndex

	//Reset the log term and index
	rl.logTerm = entry.Term
	rl.logIndex = prevLogIdx + 1

	return AppendEntryResponse{
		Term:         int64(rl.logTerm),
		LogIndex:     rl.logIndex,
		PrevLogTerm:  int64(prevLogTerm),
		PrevLogIndex: prevLogIdx,
	}, nil
}
