package term

type RaftTerm interface {
	GetTerm() int64
	IncrementTerm() int64
}
