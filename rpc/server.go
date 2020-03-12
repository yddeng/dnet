package rpc

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
)

type Server struct {
	methods   sync.Map //map[reflect.Type]*methodInfo
	lastReqNo uint64
}

type methodInfo struct {
	method   reflect.Value
	argType  reflect.Type
	respType reflect.Type
	needResp bool
}

func (server *Server) Register(method interface{}) error {
	mValue := reflect.ValueOf(method)
	if mValue.Kind() != reflect.Func {
		return fmt.Errorf("register rpc method is fail,need func(req pointer) or func(req pointer,resp pointer)")
	}

	mType := reflect.TypeOf(method)

	if mType.NumIn() < 1 || mType.NumIn() > 2 {
		return fmt.Errorf("register rpc method is fail,need func(req pointer) or func(req pointer,resp pointer)")
	}

	info := &methodInfo{
		method:  mValue,
		argType: mType.In(0),
	}

	info.needResp = mType.NumIn() == 2
	if info.needResp {
		info.respType = mType.In(1)
	}

	argType := mType.In(0)
	_, ok := server.methods.Load(argType)
	if ok {
		return fmt.Errorf("duplicate method:%s", argType)
	}
	server.methods.Store(argType, info)
	return nil
}

func (server *Server) checkMethod(name reflect.Type, needResp bool) (*methodInfo, error) {
	method, ok := server.methods.Load(name)
	if !ok {
		return nil, fmt.Errorf("not register method:%s", name.String())
	}

	m := method.(*methodInfo)
	if m.needResp != needResp {
		return nil, fmt.Errorf("invaild method:%s,register needResp=%v but request needResp=%v",
			name.String(), m.needResp, needResp)
	}
	return m, nil
}

func (server *Server) OnRPCRequest(channel RPCChannel, req *Request) error {
	// 重复请求
	if !atomic.CompareAndSwapUint64(&server.lastReqNo, req.SeqNo-1, req.SeqNo) {
		return fmt.Errorf("repeated reqNo:%d", req.SeqNo)
	}

	name := reflect.TypeOf(req.Data)
	method, err := server.checkMethod(name, req.NeedResp)
	if err != nil {
		if req.NeedResp {
			_ = channel.SendResponse(&Response{SeqNo: req.SeqNo, Err: err})
		}
		return err
	}

	var resp *Response
	arg := reflect.ValueOf(req.Data)
	if method.needResp {
		elem := reflect.New(method.respType.Elem())
		method.method.Call([]reflect.Value{arg, elem})
		resp = &Response{SeqNo: req.SeqNo, Data: elem.Interface()}
	} else {
		method.method.Call([]reflect.Value{arg})
	}

	return channel.SendResponse(resp)
}

func NewServer() *Server {
	return &Server{}
}
