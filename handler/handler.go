package handler

import (
	"github.com/golang/protobuf/proto"
	"github.com/tagDong/dnet"
	"reflect"
)

type handler func(dnet.Session, proto.Message)

type Handler struct {
	handlers map[string]handler
}

func NewHandler() *Handler {
	return &Handler{
		handlers: map[string]handler{},
	}
}

func name(msg proto.Message) string {
	return reflect.TypeOf(msg).String()
}

func (this *Handler) Register(msg proto.Message, h handler) {
	msgName := name(msg)
	if nil == h {
		return
	}
	_, ok := this.handlers[msgName]
	if ok {
		return
	}

	this.handlers[msgName] = h
}

func (this *Handler) Dispatch(session dnet.Session, msg proto.Message) {
	if nil != msg {
		name := name(msg)
		handler, ok := this.handlers[name]
		if ok {
			handler(session, msg)
		}
	}
}
