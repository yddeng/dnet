package drpc

type Request struct {
	SeqNo  uint64 // the number of request
	Method string // The name of the service and method to call.
	Data   interface{}
}

type Response struct {
	SeqNo uint64 // the number of request
	Data  interface{}
}

// RPCChannel
type RPCChannel interface {
	SendRequest(req *Request) error    // send rpc request
	SendResponse(resp *Response) error // send rpc response
}
