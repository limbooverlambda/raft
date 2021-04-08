package actors

import (
	"time"

	"kitengo/raft/log"
	"kitengo/raft/timer"
	"kitengo/raft/transport"
	"kitengo/raft/voter"
)

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
	voter.RaftVoter
}

