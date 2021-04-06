package voter

type RaftVoter interface {
	ProcessVote(vote []byte) error
}
