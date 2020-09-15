package drpc

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

type Server struct {
	methods map[string]MethodHandler
	*sync.RWMutex
}

type MethodHandler func(replyer *Replyer, req interface{})

func (server *Server) Register(name string, h MethodHandler) error {
	if name == "" {
		panic("name == ''")
	}
	if nil == h {
		panic("h == nil")
	}

	server.Lock()
	defer server.Unlock()
	_, ok := server.methods[name]
	if ok {
		return fmt.Errorf("duplicate method:%s", name)
	}
	server.methods[name] = h
	return nil
}

func (server *Server) OnRPCRequest(channel RPCChannel, req *Request) error {
	var err error
	replyer := &Replyer{Channel: channel, req: req}

	server.RLock()
	method, ok := server.methods[req.Method]
	server.RUnlock()
	if !ok {
		err = fmt.Errorf("invalid method:%s", req.Method)
		_ = replyer.reply(&Response{SeqNo: req.SeqNo, Err: err})
		return err
	}

	return server.callMethod(method, replyer)
}

func (server *Server) callMethod(method MethodHandler, replyer *Replyer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 65535)
			l := runtime.Stack(buf, false)
			err = fmt.Errorf(fmt.Sprintf("%v: %s", r, buf[:l]))
		}
	}()

	method(replyer, replyer.req.Data)
	return nil
}

type Replyer struct {
	Channel RPCChannel
	fired   int32 //防止重复Reply
	req     *Request
}

func (r *Replyer) Reply(ret interface{}, err error) error {
	if !r.req.NeedResp && atomic.CompareAndSwapInt32(&r.fired, 0, 1) {
		return fmt.Errorf("reply failde, needResp %v , fired %d", r.req.NeedResp, atomic.LoadInt32(&r.fired))
	}

	return r.reply(&Response{SeqNo: r.req.SeqNo, Data: ret, Err: err})
}

func (r *Replyer) reply(resp *Response) error {
	return r.Channel.SendResponse(resp)
}

func NewServer() *Server {
	return &Server{
		methods: map[string]MethodHandler{},
		RWMutex: new(sync.RWMutex),
	}
}
