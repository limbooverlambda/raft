package voter

type VoteStatus int

const (
	Leader VoteStatus = iota
	Split
)

type RaftVoter interface {
	RequestVote(term int64) <-chan VoteStatus
}

func NewRaftVoter() RaftVoter {
	return raftVoter{}
}

type raftVoter struct{}

func (raftVoter) RequestVote(term int64) <-chan VoteStatus {
	voteStatusChan := make(chan VoteStatus, 1)
	go func() {
		defer close(voteStatusChan)
		voteStatusChan <- Leader
	}()
	return voteStatusChan
}
