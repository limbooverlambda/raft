package term

type RaftTerm interface {
	GetTerm() int64
	IncrementTerm() int64
}

func NewRaftTerm() RaftTerm {
	return raftTerm{}
}

type raftTerm struct{}

func (raftTerm) GetTerm() int64 {
	panic("implement me")
}

func (raftTerm) IncrementTerm() int64 {
	panic("implement me")
}
