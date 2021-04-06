package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"kitengo/raft/actors"
)

func main() {
	fmt.Println("Starting raft...")
	ctx := context.Background()
	cancellableCtxt, cancel := context.WithCancel(ctx)
	candidateChan := make(chan actors.CandidateTrigger)
	followerChan := make(chan actors.FollowerTrigger)
	leaderChan := make(chan actors.LeaderTrigger)
	follower := actors.NewFollower(cancellableCtxt, followerChan, candidateChan)
	candidate := actors.NewCandidate(cancellableCtxt, leaderChan, candidateChan)
	leader := actors.NewLeader(cancellableCtxt, leaderChan)

	follower.Run()
	candidate.Run()
	leader.Run()
	followerChan <- actors.FollowerTrigger{}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<- sig
	cancel()

}

