package actors

type AppendEntryRequest struct{}
type AppendEntryResponse struct{}

type RVoteRequest struct{}
type RVoteResponse struct{}

type RaftActor interface {
	AppendEntry(appendEntryRequest AppendEntryRequest) (AppendEntryResponse, error)
	RequestVote(rVoteRequest RVoteRequest) (RVoteResponse, error)
}

type raftActor struct {
	aeRequestChan     chan<- AppendEntryRequest
	aeResponseChan    <-chan AppendEntryResponse
	aeErrorChan       <-chan error
	rVoteRequestChan  chan<- RVoteRequest
	rVoteResponseChan <-chan RVoteResponse
	rVoteErrorChan    <-chan error
}

func (ra *raftActor) AppendEntry(appendEntryRequest AppendEntryRequest) (AppendEntryResponse, error) {
	panic("!!!")
}

func (ra *raftActor) RequestVote(rVoteRequest RVoteRequest) (RVoteResponse, error) {
	panic("implement me")
}
