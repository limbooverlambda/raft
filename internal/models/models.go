package models

import (
	"bytes"
	"encoding/gob"
)

//RequestType defines the kind of request received
//by the RaftServer, it can be one of 'AppendEntry'
// 'RequestVote' or 'ClientCommand'
//go:generate stringer -type=RequestType
type RequestType int

const (
	AppendEntry RequestType = iota
	RequestVote
	ClientCommand
)

type Request struct {
	RequestType RequestType
	Payload     []byte
}

type RequestConverter interface {
	ToRequest() (Request, error)
}

type AppendEntryPayload struct {
	Term         int64
	LeaderId     string
	PrevLogIndex int64
	PrevLogTerm  int64
	Entries      []byte
	LeaderCommit int64
}

func (aep *AppendEntryPayload) ToRequest() (Request, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(aep);err != nil {
		return Request{}, err
	}
	return Request{
		RequestType: AppendEntry,
		Payload:    buffer.Bytes(),
	}, nil
}

type RequestVotePayload struct {
	Term         int64
	CandidateId  string
	LastLogIndex int64
	LastLogTerm  int64
}

type ClientCommandPayload struct {
	ClientCommand []byte
}

func (ccp *ClientCommandPayload) ToRequest() (Request, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(ccp);err != nil {
		return Request{}, err
	}
	return Request{
		RequestType: ClientCommand,
		Payload:    buffer.Bytes(),
	}, nil
}

type Response struct {
	Payload []byte
}
