package elector

type ElectionStatusChan <-chan bool

type RaftElector interface {
	StartElection() ElectionStatusChan
	StopElection()
}
