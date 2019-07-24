package message

import (
	"reflect"
)

type Message struct {
	name     string
	serialNO uint16
	data     interface{}
}

func NewMessage(seriNo uint16, data interface{}) *Message {
	return &Message{
		serialNO: seriNo,
		data:     data,
	}
}

func (m *Message) GetSerialNo() uint16 {
	return m.serialNO
}

func (m *Message) GetData() interface{} {
	return m.data
}

func (m *Message) GetName() string {
	return reflect.TypeOf(m.data).String()
}
