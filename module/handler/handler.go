package handler

import (
	"github.com/tagDong/dnet"
	"reflect"
)

type handler func(dnet.Session, dnet.Message)

type Handler struct {
	handlers map[string]handler
}

func NewHandler() *Handler {
	return &Handler{
		handlers: map[string]handler{},
	}
}

func (this *Handler) RegisterCallBack(descriptor interface{}, callback func(session dnet.Session, msg dnet.Message)) {
	msgName := reflect.TypeOf(descriptor).String()
	if nil == callback {
		return
	}
	_, ok := this.handlers[msgName]
	if ok {
		return
	}

	this.handlers[msgName] = callback
}

func (this *Handler) Dispatch(session dnet.Session, msg dnet.Message) {
	if nil != msg {
		name := msg.GetName()
		handler, ok := this.handlers[name]
		if ok {
			handler(session, msg)
		}
	}
}
