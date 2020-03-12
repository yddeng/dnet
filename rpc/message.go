package rpc

type Request struct {
	SeqNo    uint64
	Data     interface{}
	NeedResp bool
}

type Response struct {
	SeqNo uint64
	Data  interface{}
	Err   error
}

type RPCChannel interface {
	SendRequest(req *Request) error    // 发送RPC请求
	SendResponse(resp *Response) error // 发送RPC回复
}
