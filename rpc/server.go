package rpc

import (
	"fmt"
	"log"
	"reflect"
	"sync"
)

type Server struct {
	severCodec ServerCodec
	methods    sync.Map //map[reflect.Type]*methodInfo
}

type methodInfo struct {
	method   reflect.Value
	argType  reflect.Type
	respType reflect.Type
	needResp bool
}

func (server *Server) Register(method interface{}) {
	mValue := reflect.ValueOf(method)
	if mValue.Kind() != reflect.Func {
		log.Fatalln("register rpc method is fail,need func(req pointer) or func(req pointer,resp pointer)")
	}

	mType := reflect.TypeOf(method)

	if mType.NumIn() < 1 || mType.NumIn() > 2 {
		log.Fatalln("register rpc method is fail,need func(req pointer) or func(req pointer,resp pointer)")
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
		log.Printf("duplicate method:%s", argType)
	}
	server.methods.Store(argType, info)
}

func (server *Server) checkMethod(name reflect.Type, needResp bool) (*methodInfo, error) {
	method, ok := server.methods.Load(name)
	if !ok {
		return nil, fmt.Errorf("not register method:%s", name.String())
	}

	m := method.(*methodInfo)
	if m.needResp != needResp {
		return nil, fmt.Errorf("invaild method:%s ,register needResp = %v : request needResp = %v",
			name.String(), m.needResp, needResp)
	}
	return m, nil
}

func (server *Server) reply(channel RPCChannel, resp *Response) {
	data, err := server.severCodec.EncodeResponse(resp)
	if err != nil {
		log.Println(err)
		return
	}
	err = channel.SendResponse(data)
	if err != nil {
		log.Println(err)
	}
}

func (server *Server) OnRPCRequest(channel RPCChannel, data interface{}) {

	req, err := server.severCodec.DecodeRequest(data)
	if err != nil {
		log.Println(err)
		return
	}

	name := reflect.TypeOf(req.Data)
	method, err := server.checkMethod(name, req.NeedResp)
	if err != nil {
		server.reply(channel, &Response{SeqNo: req.SeqNo, Err: err})
		log.Println(err)
		return
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

	server.reply(channel, resp)
}

func NewServer(severCodec ServerCodec) *Server {
	return &Server{
		severCodec: severCodec,
	}
}
