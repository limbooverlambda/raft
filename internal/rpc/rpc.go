package rpc

type RaftRpc interface {
	Receive(request RaftRpcRequest)
	Process(meta RaftRpcMeta) (RaftRpcResponse, error)
	RaftRpcReqChan() <-chan RaftRpcRequest
}

type RaftRpcRequest interface {
	GetResponseChan() chan<- RaftRpcResponse
	GetErrorChan() chan<- error
}
type RaftRpcMeta interface{}
type RaftRpcResponse interface{}
