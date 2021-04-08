package actors

import (
	"time"

	"kitengo/raft/log"
	"kitengo/raft/term"
	"kitengo/raft/timer"
	"kitengo/raft/transport"
	"kitengo/raft/voter"
)

type candidateStub struct {
	fakeTransport fakeTransport
	fakeTimer     fakeTimer
	fakeLog       fakeLog
	fakeVoter     fakeVoter
	fakeTerm      fakeTerm
	leaderTrigger chan<- LeaderTrigger
}

type fakeTransport struct {
	GetRequestChanFn func() <-chan transport.Request
	SendResponseFn func(response transport.Response) error
	GetRequestTypeFn func(req transport.Request) transport.RequestType
	transport.Transport
}

func (f fakeTransport) GetRequestChan() <-chan transport.Request {
	return f.GetRequestChanFn()
}

func (f fakeTransport) SendResponse(response transport.Response) error {
	return f.SendResponseFn(response)
}

func (f fakeTransport) GetRequestType(req transport.Request) transport.RequestType {
	return f.GetRequestTypeFn(req)
}

type fakeTimer struct {
	SetDeadlineFn func(time.Time)
	GetDeadlineFn func() time.Time
	timer.RaftTimer
}

func (ft fakeTimer) SetDeadline(currentTime time.Time) {
	ft.SetDeadlineFn(currentTime)
}

func (ft fakeTimer) GetDeadline() time.Time {
	return ft.GetDeadlineFn()
}

type fakeLog struct {
	AppendEntryFn func(entry []byte)
	log.RaftLog
}

func (fl fakeLog) AppendEntry(entry []byte) {
	fl.AppendEntryFn(entry)
}

type fakeVoter struct {
	RequestVoteFn func(term int64) <-chan voter.VoteStatus
	voter.RaftVoter
}

func (fv fakeVoter) RequestVote(term int64) <-chan voter.VoteStatus {
	return fv.RequestVoteFn(term)
}

type fakeTerm struct {
	term.RaftTerm
	GetTermFn func() int64
	IncrementTermFn func() int64
}

func (fterm fakeTerm) GetTerm() int64 {
	return fterm.GetTermFn()
}

func (fterm fakeTerm) IncrementTerm() int64 {
	return fterm.IncrementTermFn()
}

