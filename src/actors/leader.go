package actors

import (
	"context"
	"fmt"
	"time"

	"kitengo/raft/rconfig"
)

type Leader interface {
	Run()
}

type LeaderTrigger struct {}

func NewLeader(ctx context.Context,
	leaderTrigger <- chan LeaderTrigger) Leader {
	return &leader{
		context: ctx,
		leaderTrigger: leaderTrigger,
	}
}

type leader struct {
	context context.Context
	leaderTrigger <- chan LeaderTrigger
}

func (f *leader) Run() {
	go func() {
		ticker := time.NewTicker(rconfig.PollDuration * time.Second)
		defer ticker.Stop()
		fmt.Println("Leader running")
		for range ticker.C {
			fmt.Println("Leader ticking")
			select {
			case <-f.context.Done():
				fmt.Println("Cancelling Leader")
				return
			case <-f.leaderTrigger:
				fmt.Println("Leader triggered")
			default:
				fmt.Println("Leader ticking...")
			}
		}
	}()
}

