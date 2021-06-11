package log

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/kitengo/raft/internal/rafterror"
	"io"
	"log"
	"os"
	"sync"
)

var (
	byteorder = binary.BigEndian
)

const (
	ErrEmptyPayloadAppend = rafterror.Error("cannot append empty payload")
	ErrIllegalIndex       = rafterror.Error("cannot read from index")
	ErrReadingFromLog     = rafterror.Error("unable to read from log file")
	ErrReadingFromIndex   = rafterror.Error("unable to read from index file")
	ErrWritingToLog       = rafterror.Error("unable to write to log file")
)

const (
	recordLengthInBytes   = 8
	metadataLengthInBytes = 24
)

type Entry struct {
	Term    uint64
	Payload []byte
}

type EntryMeta struct {
	Term         uint64
	LogIndex     uint64
	PrevLogTerm  uint64
	PrevLogIndex uint64
	PayloadSize  uint64
}

type Index struct {
	Position    uint64
	Term        uint64
	PayloadSize uint64
}

//RaftLog is an abstraction around the index and the raw commands to the
//replicated state machine.
// - Create an index file
// - Create the raw log file
type RaftLog interface {
	//AppendEntry is used to append the payload and term to the log
	AppendEntry(entry Entry) (EntryMeta, error)
	//LastLogEntryMeta retrieves the metadata for the last log entry
	LastLogEntryMeta() (EntryMeta, error)
	//LogEntry retrieves the Entry appended to the log
	LogEntry(indexID uint64) (Entry, error)
	//LogEntryMeta retrieves the metadata at the provided index
	LogEntryMeta(indexID uint64) (EntryMeta, error)
	//Truncate remove all the index entries including and following provided indexID
	Truncate(indexID uint64) error
	//GetLogEntryMetaAtIndex(index uint64) (EntryMeta, error)
	//GetLogEntryAtIndex(index uint64) (Entry, error)
	//TruncateFromIndex(index uint64) error
	//GetCurrentLogEntry() EntryMeta
	//GetCurrentLogIndex() uint64
}

func NewRaftLog(logID string) RaftLog {
	return initRaftLog(logID)
}

func initRaftLog(logID string) *raftLog {
	logFile, logSize, err := openFile(logID)
	if err != nil {
		panic(fmt.Errorf("unable to open log file due to %v", err))
	}
	idxFile, idxSize, err := openFile(logID + "_idx")
	if err != nil {
		panic(fmt.Errorf("unable to open index file due to %v", err))
	}
	return &raftLog{
		logID:       logID,
		logFile:     logFile,
		logBuf:      bufio.NewWriter(logFile),
		idxFile:     idxFile,
		idxBuf:      bufio.NewWriter(idxFile),
		idxSize:     uint64(idxSize),
		logSize:     uint64(logSize),
		logPosition: 0,
		logTerm:     0,
		logIndex:    0,
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

//This is called only from the test at the moment
func deleteFile(id string) error {
	if err := os.Remove(id); err != nil {
		return err
	}
	return os.Remove(id + "_idx")

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

func (rl *raftLog) AppendEntry(entry Entry) (EntryMeta, error) {
	rl.Lock()
	defer rl.Unlock()
	if len(entry.Payload) == 0 {
		return EntryMeta{}, ErrEmptyPayloadAppend
	}
	logPosition := rl.logSize
	entrySize, err := rl.writeToLog(entry.Payload)
	if err != nil {
		return EntryMeta{}, err
	}
	rl.logSize += uint64(entrySize)
	//Add the position, term and payload size to the index
	index := Index{
		Position:    logPosition,
		Term:        entry.Term,
		PayloadSize: uint64(entrySize),
	}
	_, err = rl.writeToIndex(index)
	if err != nil {
		return EntryMeta{}, err
	}

	//prevLogTerm and prevLogIndex
	prevLogTerm := rl.logTerm
	prevLogIdx := rl.logIndex

	//Reset the log term and index
	rl.logTerm = entry.Term
	rl.logIndex = prevLogIdx + 1
	rl.idxSize = rl.idxSize + metadataLengthInBytes

	return EntryMeta{
		Term:         rl.logTerm,
		LogIndex:     rl.logIndex,
		PrevLogTerm:  prevLogTerm,
		PrevLogIndex: prevLogIdx,
		PayloadSize:  uint64(entrySize),
	}, nil
}

func (rl *raftLog) Truncate(indexID uint64) error {
	idxTruncationSize := rl.idxSize - (indexID-1)*metadataLengthInBytes
	if err := rl.idxFile.Truncate(int64(idxTruncationSize)); err != nil {
		return err
	}
	rl.idxSize = idxTruncationSize
	rl.logIndex = indexID - 1
	return nil
}

func (rl *raftLog) LogEntryMeta(index uint64) (EntryMeta, error) {
	rl.RLock()
	defer rl.RUnlock()
	idx, err := rl.readFromIndex(index)
	if err != nil {
		return EntryMeta{}, err
	}
	return EntryMeta{
		Term:         idx.Term,
		LogIndex:     index,
		PrevLogTerm:  0,
		PrevLogIndex: 0,
		PayloadSize:  idx.PayloadSize,
	}, err
}

func (rl *raftLog) LogEntry(indexID uint64) (Entry, error) {
	rl.RLock()
	defer rl.RUnlock()
	idx, err := rl.readFromIndex(indexID)
	if err != nil {
		return Entry{}, ErrReadingFromIndex
	}
	payload, err := rl.readFromLog(idx.Position, idx.PayloadSize)
	if err != nil {
		return Entry{}, ErrReadingFromLog
	}
	return Entry{
		Term:    idx.Term,
		Payload: payload,
	}, nil
}

func (rl *raftLog) LastLogEntryMeta() (EntryMeta, error) {
	rl.RLock()
	defer rl.RUnlock()
	lastIndex := rl.idxSize / metadataLengthInBytes
	idx, err := rl.readFromIndex(lastIndex)
	if err != nil {
		return EntryMeta{}, err
	}
	return EntryMeta{
		Term:         idx.Term,
		LogIndex:     lastIndex,
		PrevLogTerm:  0,
		PrevLogIndex: 0,
		PayloadSize:  idx.PayloadSize,
	}, err
}


func (rl *raftLog) writeToLog(payload []byte) (entrySize int, err error) {
	if entrySize, err = rl.logBuf.Write(payload); err != nil {
		return 0, ErrWritingToLog
	}
	if err = rl.logBuf.Flush(); err != nil {
		return entrySize, ErrWritingToLog
	}
	return entrySize, nil
}

func (rl *raftLog) readFromLog(readPosition uint64, payloadSize uint64) ([]byte, error) {
	payloadBytes := make([]byte, int64(payloadSize))
	_, err := rl.logFile.ReadAt(payloadBytes, int64(readPosition))
	if err != nil && err != io.EOF {
		return nil, ErrReadingFromLog
	}
	return payloadBytes, nil
}

func (rl *raftLog) readFromIndex(indexID uint64) (Index, error) {
	metadataBytes := make([]byte, metadataLengthInBytes)
	offset := (indexID - 1) * metadataLengthInBytes
	_, err := rl.idxFile.ReadAt(metadataBytes, int64(offset))
	if err != nil && err != io.EOF {
		return Index{}, ErrIllegalIndex
	}
	logPosition, logTerm, logEntrySize := byteorder.Uint64(metadataBytes[0:recordLengthInBytes]),
		byteorder.Uint64(metadataBytes[recordLengthInBytes:recordLengthInBytes+8]),
		byteorder.Uint64(metadataBytes[recordLengthInBytes+8:recordLengthInBytes+16])
	return Index{
		Position:    logPosition,
		Term:        logTerm,
		PayloadSize: logEntrySize,
	}, nil
}

func (rl *raftLog) writeToIndex(index Index) (Index, error) {
	if err := binary.Write(rl.idxBuf, byteorder, index.Position); err != nil {
		log.Printf("Failed to write log position to index %v\n", err)
		return Index{}, err
	}

	if err := binary.Write(rl.idxBuf, byteorder, index.Term); err != nil {
		log.Printf("Failed to write entry term to index %v\n", err)
		return Index{}, err
	}
	//Need to cast to int32 since binary.Write will fail on data that doesn't have a fixed size
	payloadSize := index.PayloadSize
	if err := binary.Write(rl.idxBuf, byteorder, payloadSize); err != nil {
		log.Printf("Failed to write payload size to index %v\n", err)
		return Index{}, err
	}

	//Flush to the disc right away
	err := rl.idxBuf.Flush()
	if err != nil {
		log.Printf("Failed to flush indexbuffer payload to disc %v\n", err)
		return Index{}, err
	}
	return index, nil
}
