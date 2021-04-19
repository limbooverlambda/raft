package state

type State int

const (
	FollowerState State = iota
	CandidateState
	LeaderState
	ShutdownState
)

type RaftState interface {
	GetStateChan() <- chan State
	SetState(state State)
}

type raftState struct {
	stateChan chan State
}

func NewRaftState() RaftState {
    stateChan :=  make(chan State, 1)
	//Seed the FollowerState
    stateChan <- FollowerState
    return &raftState{
		stateChan: stateChan,
	}
}

func (rs *raftState) GetStateChan() <- chan State {
	return rs.stateChan
}

func (rs *raftState) SetState(state State) {
	rs.stateChan <- state
}


