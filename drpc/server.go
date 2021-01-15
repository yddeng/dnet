package drpc

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

type Server struct {
	methods map[string]MethodHandler
	mtx     sync.RWMutex
}

type MethodHandler func(replyer *Replyer, req interface{})

func (server *Server) Register(name string, h MethodHandler) {
	if name == "" {
		panic("drpc: Register name == ''")
	}
	if nil == h {
		panic("drpc: Register h == nil")
	}

	server.mtx.Lock()
	defer server.mtx.Unlock()
	_, ok := server.methods[name]
	if ok {
		panic(fmt.Sprintf("drpc: Register duplicate method:%s", name))
	}
	server.methods[name] = h
}

func (server *Server) OnRPCRequest(channel RPCChannel, req *Request) error {
	if channel == nil || req == nil {
		return fmt.Errorf("drpc: OnRPCRequest invalid argument")
	}

	server.mtx.RLock()
	method, ok := server.methods[req.Method]
	server.mtx.RUnlock()
	if !ok {
		return fmt.Errorf("drpc: OnRPCRequest invalid method %s", req.Method)
	}

	replyer := &Replyer{Channel: channel, req: req}
	return server.callMethod(method, replyer)
}

func (server *Server) callMethod(method MethodHandler, replyer *Replyer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1024)
			l := runtime.Stack(buf, false)
			err = fmt.Errorf(fmt.Sprintf("%v: %s", r, buf[:l]))
		}
	}()

	method(replyer, replyer.req.Data)
	return
}

type Replyer struct {
	Channel RPCChannel
	fired   int32 //防止重复Reply
	req     *Request
}

func (r *Replyer) Reply(ret interface{}) error {
	if ret == nil {
		return fmt.Errorf("drpc: Reply ret == nil")
	}

	if !atomic.CompareAndSwapInt32(&r.fired, 0, 1) {
		return fmt.Errorf("drpc: Reply repeated reply %d ", atomic.LoadInt32(&r.fired))
	}

	return r.reply(&Response{SeqNo: r.req.SeqNo, Data: ret})
}

func (r *Replyer) reply(resp *Response) error {
	return r.Channel.SendResponse(resp)
}

func NewServer() *Server {
	return &Server{
		methods: map[string]MethodHandler{},
	}
}
