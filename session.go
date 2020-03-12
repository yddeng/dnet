package dnet

import (
	"io"
	"net"
	"time"
)

type Session interface {
	/*
	  编解码器
	  Start之前需设置编解码器，否则使用默认的编解码器
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
	 回掉函数返回 message，err。当且仅当 err == nil，message 不为空。
	 返回错误信息后没有主动关闭连接，需要主动调用Close。（io.EOF、编解码错误）
	*/
	Start(func(message interface{}, err error)) error
	// 读写超时
	SetTimeout(readTimeout, writeTimeout time.Duration)
	// 发送一个对象，经过编码发送出去
	Send(o interface{}) error
	// 发送数据，不经过编码器直接发送
	SendBytes(data []byte) error
	// 给session绑定用户数据
	SetUserData(ud interface{})
	// 获取用户数据
	GetUserData() interface{}
	// 连接断开回调
	SetCloseCallBack(func(reason string))
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
