package rpc

type AppendEntryRequest struct{}
type AppendEntryResponse struct {
	Success bool
}

type RaftAppendEntry interface {
	AppendEntryReqChan() <- chan AppendEntryRequest
	Process(request AppendEntryRequest) AppendEntryResponse
	AppendEntryRespChan() chan <- AppendEntryResponse
}
