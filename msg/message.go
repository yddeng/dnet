package msg

import (
	"github.com/golang/protobuf/proto"
	"reflect"
)

type Message struct {
	data proto.Message
}

func NewMessage(data proto.Message) *Message {
	return &Message{data: data}
}

func (this *Message) GetData() proto.Message {
	return this.data
}

func (this *Message) GetName() string {
	return reflect.TypeOf(this.data).String()
}
