package handler

import (
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/examples/module/message"
	"reflect"
)

/*
 事件分发器
 每一条协议注册一个处理函数
*/

type handler func(dnet.Session, *message.Message)

type Handler struct {
	handlers map[string]handler
}

func NewHandler() *Handler {
	return &Handler{
		handlers: map[string]handler{},
	}
}

func (this *Handler) RegisterCallBack(descriptor interface{}, callback func(session dnet.Session, msg *message.Message)) {
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

func (this *Handler) Dispatch(session dnet.Session, msg *message.Message) {
	if nil != msg {
		name := msg.GetName()
		handler, ok := this.handlers[name]
		if ok {
			handler(session, msg)
		}
	}
}
