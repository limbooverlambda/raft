package voter

type VoteStatus int

const (
	Leader VoteStatus = iota
	Split
)

type RaftVoter interface {
	ProcessVote(vote []byte) error
	RequestVote(term int64) <- chan VoteStatus
}
