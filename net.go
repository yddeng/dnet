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

// NetConn
type NetConn interface {
	net.Conn
}

// AcceptorHandle type interface
type AcceptorHandle interface {
	// handler to invokes
	OnConnection(conn NetConn)
}

type AcceptorHandlerFunc func(conn NetConn)

func (handler AcceptorHandlerFunc) OnConnection(conn NetConn) {
	// handler to invokes
	handler(conn)
}

// HandleFunc returns AcceptorHandlerFunc with the handler function.
func HandleFunc(handler func(conn NetConn)) AcceptorHandlerFunc {
	return handler
}

// Acceptor type interface
type Acceptor interface {
	// Serve listen and serve
	Serve(handler AcceptorHandle) error

	// Stop stop the acceptor
	Stop()

	// Addr returns address of the listener
	Addr() net.Addr
}

// Codec
type Codec interface {
	// Encode
	Encode(o interface{}) ([]byte, error)

	// Decode
	Decode(reader io.Reader) (interface{}, error)
}