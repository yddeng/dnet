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

const (
	defSendChannelSize = 1024
)

type Session interface {
	// connection
	NetConn() interface{}

	// RemoteAddr returns the remote network address.
	RemoteAddr() net.Addr

	// LocalAddr returns the local network address.
	LocalAddr() net.Addr

	// Send data will be encoded by the encoder and sent
	Send(o interface{}) error

	// SetContext binding session data
	SetContext(ctx interface{})

	// Context returns binding session data
	Context() interface{}

	// Close closes the session.
	Close(reason error)

	// IsClosed returns has it been closed
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
