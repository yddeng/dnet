package dnet

import (
	"io"
	"net"
	"time"
)

type Session interface {
	/*
	  编解码器
	  用于消息发送编码，接收解码
	  需要在 Start 前设置
	*/
	SetCodec(codec Codec)

	// conn
	NetConn() interface{}

	// 获取远端地址
	RemoteAddr() net.Addr

	// 获取远端地址
	LocalAddr() net.Addr

	/*
	 开启数据接收处理
	 callback 函数返回 message，err。当且仅当 err == nil 时，message 不为空。
	 返回错误信息后没有主动关闭连接，需要主动调用Close。（io.EOF、编解码错误）
	*/
	Start(callback func(message interface{}, err error)) error

	// 读写超时
	SetTimeout(readTimeout, writeTimeout time.Duration)

	// 发送一个对象，经过编码发送出去
	Send(o interface{}) error

	// 发送数据，不经过编码器直接发送
	SendBytes(data []byte) error

	// 给session绑定用户数据
	SetContext(ctx interface{})

	// 获取用户数据
	Context() interface{}

	// 连接断开回调
	SetCloseCallBack(func(session Session, reason string))

	/*
	 主动关闭连接
	 先关闭读，待数据发送完毕再关闭连接
	*/
	Close(reason string)
}

//编解码器
type Codec interface {
	//编码
	Encode(interface{}) ([]byte, error)
	//解码
	Decode(reader io.Reader) (interface{}, error)
}
