package dnet

import (
	"io"
)

//编解码器
type Codec interface {
	//编码
	Encode(interface{}) ([]byte, error)
	//解码
	Decode(reader io.Reader) (interface{}, error)
}

//消息分发器
type Dispatcher interface {
	//注册
	RegisterCallBack(descriptor interface{}, callback func(session Session, msg Message))
	//分发
	Dispatch(session Session, msg Message)
}

//用于传输
type Message interface {
	GetData() interface{}
	//获取序列号
	GetSerialNo() uint16
	GetName() string
}
