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
	SendRequest(interface{}) error  // 发送RPC请求
	SendResponse(interface{}) error // 发送RPC回复
}

type ServerCodec interface {
	//编码
	EncodeResponse(response *Response) ([]byte, error)
	//解码
	DecodeRequest(data interface{}) (*Request, error)
}

type ClientCodec interface {
	//编码
	EncodeRequest(request *Request) ([]byte, error)
	//解码
	DecodeResponse(data interface{}) (*Response, error)
}
