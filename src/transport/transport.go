package transport

type Request struct{
	Payload []byte
}
type Response struct{}

type RequestType int

const (
	AppendEntry RequestType = iota
	VoteRequest
)

type Transport interface {
	GetRequestChan() <-chan Request
	SendResponse(response Response) error
	GetRequestType(req Request) RequestType
}
