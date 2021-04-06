package actors

import (
	"context"
	"fmt"
	"time"

	"kitengo/raft/rconfig"
)

type Follower interface {
	Run()
}

type FollowerTrigger struct{}

func NewFollower(ctx context.Context,
	followerTrigger <- chan FollowerTrigger,
	candidateTrigger chan <- CandidateTrigger,
	) Follower {
	return &follower{
		context: ctx,
		followerTrigger: followerTrigger,
		candidateTrigger: candidateTrigger,
	}
}

type follower struct {
	context context.Context
	followerTrigger <- chan FollowerTrigger
	candidateTrigger chan <- CandidateTrigger
}

func (f *follower) Run() {
	go func() {
		fmt.Println("Follower running..")
		ticker := time.NewTicker(rconfig.PollDuration * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			fmt.Println("Follower ticking")
			select {
			case <-f.context.Done():
				fmt.Println("Cancelling Follower")
				return
			case <- f.followerTrigger:
				fmt.Println("Follower triggered")
				f.candidateTrigger <-  CandidateTrigger{}
			default:
				fmt.Println("Follower ticking...")
			}
		}
	}()
}
