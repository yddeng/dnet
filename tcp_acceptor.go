package dnet

import (
	"errors"
	"net"
	"sync/atomic"
	"time"
)

type TCPAcceptor struct {
	address  string
	listener net.Listener
	started  int32
}

// NewTCPAcceptor returns a new instance of TCPAcceptor
func NewTCPAcceptor(address string) *TCPAcceptor {
	return &TCPAcceptor{address: address}
}

// ServeTCP listen and serve tcp address with handler
func ServeTCP(address string, handler AcceptorHandle) error {
	return NewTCPAcceptor(address).Serve(handler)
}

// Serve listens and serve in the specified addr
func (this *TCPAcceptor) Serve(handler AcceptorHandle) error {
	if handler == nil {
		return errors.New("dnet:Serve handler is nil. ")
	}

	if !atomic.CompareAndSwapInt32(&this.started, 0, 1) {
		return errors.New("dnet:Serve acceptor is already started. ")
	}

	listener, err := net.Listen("tcp", this.address)
	if err != nil {
		return err
	}
	this.listener = listener
	defer this.Stop()

	for {
		conn, err := this.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(time.Millisecond * 5)
				continue
			} else {
				return err
			}
		}

		go handler.OnConnection(conn)
	}

}

// Addr returns the addr the acceptor will listen on
func (this *TCPAcceptor) Addr() net.Addr {
	return this.listener.Addr()
}

// Stop stops the acceptor
func (this *TCPAcceptor) Stop() {
	if atomic.CompareAndSwapInt32(&this.started, 1, 0) {
		_ = this.listener.Close()
	}
}

// DialTCP
func DialTCP(address string, timeout time.Duration) (NetConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{Timeout: timeout}
	return dialer.Dial(tcpAddr.Network(), address)
}
