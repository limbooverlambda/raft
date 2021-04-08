package transport

type Request struct{
	Payload []byte
}
type Response struct{}

type RequestType int

type AppendEntryRequest struct {}
type PeerVoteRequest struct {}

const (
	AppendEntry RequestType = iota
	VoteRequest
)

type Transport interface {
	GetRequestChan() <-chan Request
	SendResponse(response Response) error
	GetRequestType(req Request) RequestType
	AppendEntryChan() <- chan AppendEntryRequest
	VoteRequestChan() <- chan PeerVoteRequest
}
