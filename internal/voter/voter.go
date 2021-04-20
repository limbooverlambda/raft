package voter

type VoteStatus int

const (
	Leader VoteStatus = iota
	Split
)

type RaftVoter interface {
	ProcessVote(vote []byte) error
	RequestVote(term int64) <-chan VoteStatus
}

func NewRaftVoter() RaftVoter {
	return raftVoter{}
}

type raftVoter struct{}

func (raftVoter) ProcessVote(vote []byte) error {
	panic("implement me")
}

func (raftVoter) RequestVote(term int64) <-chan VoteStatus {
	panic("implement me")
}
