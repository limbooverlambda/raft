package actors

import (
	"context"
	"testing"
	"time"

	"kitengo/raft/transport"
)

func Test_follower_Run_WithCancellation(t *testing.T) {
	ctx := context.Background()
	cc, cancel := context.WithCancel(ctx)
	followerTrigger := make(<-chan FollowerTrigger)
	candidateTrigger := make(chan<- CandidateTrigger)
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
	<-c
}

func Test_follower_Run_WithExecutionAndEventualTimeout(t *testing.T) {
	ctx := context.Background()
	cc, _ := context.WithCancel(ctx)
	followerTrigger := make(chan FollowerTrigger)
	candidateTrigger := make(chan<- CandidateTrigger)
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
	<-c
}
