package dnet

import (
	"net"
)

type Session interface {
	// options
	//WithOptions(options ...Option)

	// conn
	NetConn() interface{}

	// 获取远端地址
	RemoteAddr() net.Addr

	// 获取远端地址
	LocalAddr() net.Addr

	/*
	 开启数据接收处理
	 callback 函数返回 message，err。当且仅当 err == nil 时，message 不为空。
	 返回错误信息后没有主动关闭连接，需要主动调用Close。（io.EOF、编解码错误）。
	 有解码器，经过解码器解码后返回。没有解码器，直接返回数据。
	*/
	//Start(callback func(message interface{}, err error)) error

	/*
	 * 发送
	 * 有编码器，任何数据都将经过编码器编码后发送
	 * 没有编码器，仅能发送 []byte 类型数据。
	 */
	Send(o interface{}) error

	// 给session绑定用户数据
	SetContext(ctx interface{})

	// 获取用户数据
	Context() interface{}

	/*
	 主动关闭连接
	 先关闭读，待数据发送完毕再关闭连接
	*/
	Close(reason error)
}
