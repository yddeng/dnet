package dnet

import (
	"errors"
	"io"
	"net"
)

var (
	ErrSessionClosed  = errors.New("dnet: session is closed. ")
	ErrNilMsgCallBack = errors.New("dnet: session without msgCallback")
	ErrSendMsgNil     = errors.New("dnet: session send msg is nil")
	ErrSendChanFull   = errors.New("dnet: session send channel is full")

	ErrSendTimeout = errors.New("dnet: send timeout. ")
	ErrReadTimeout = errors.New("dnet: read timeout. ")
)

var (
	ErrRPCTimeout = errors.New("dnet: rpc timeout")
)

const (
	sendBufChanSize = 1024
)

type Session interface {
	// conn
	NetConn() interface{}

	// 获取远端地址
	RemoteAddr() net.Addr

	// 获取远端地址
	LocalAddr() net.Addr

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
	 先关闭读，待数据发送完毕再关闭连接
	*/
	Close(reason error)

	// 是否已经关闭
	IsClosed() bool
}

// callback of new session
type newSessionCallback func(session Session)

// Acceptor type interface
type Acceptor interface {
	Listen(conn newSessionCallback) error
	Stop()
	Addr() net.Addr
}

//编解码器
type Codec interface {
	Encode(o interface{}) ([]byte, error)
	Decode(reader io.Reader) (interface{}, error)
}
