package actors

import (
	"context"
	"testing"
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

func Test_follower_Run_WithCancellation(t *testing.T) {
	ctx := context.Background()
	cc, cancel := context.WithCancel(ctx)
	followerTrigger := make(<- chan FollowerTrigger)
	candidateTrigger := make(chan <- CandidateTrigger)
	f := NewFollower(cc,
		followerTrigger,
		candidateTrigger,
		fakeTransport{},
		fakeTimer{},
		fakeLog{},
		fakeVoter{},
		)
	c := f.Run()
	time.Sleep(3 * time.Second)
	cancel()
	<- c
}

func Test_follower_Run_WithExecutionAndEventualTimeout(t *testing.T) {
	ctx := context.Background()
	cc, _ := context.WithCancel(ctx)
	followerTrigger := make(chan FollowerTrigger)
	candidateTrigger := make(chan <- CandidateTrigger)
	transportChan := make(chan transport.Request)

	transpo := fakeTransport{}
	transpo.GetRequestChanFn = func() <-chan transport.Request {
		return transportChan
	}
	transpo.GetRequestTypeFn = func(req transport.Request) transport.RequestType {
		return transport.AppendEntry
	}

	rLog := fakeLog{}
	rLog.AppendEntryFn = func(entry []byte) {
		//NO-OP
	}

	rTimer := fakeTimer{}
	rTimer.SetDeadlineFn = func(t time.Time) {
		//NO-OP
	}
	rTimer.GetDeadlineFn = func() time.Time {
		//Timeout the execution right now
		return time.Now()
	}
	rVoter := fakeVoter{}

	f := NewFollower(cc,
		followerTrigger,
		candidateTrigger,
		transpo,
		rTimer,
		rLog,
		rVoter)

	c := f.Run()
	followerTrigger <- FollowerTrigger{}
	transportChan <- transport.Request{}
	<- c
}


