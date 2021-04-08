package dnet

import (
	"errors"
	"net"
	"sync/atomic"
	"time"
)

type TCPAcceptor struct {
	tcpAddr  *net.TCPAddr
	listener *net.TCPListener
	options  []Option
	started  int32
}

// NewTCPAcceptor returns a new instance of TCPAcceptor
func NewTCPAcceptor(address string, options ...Option) (*TCPAcceptor, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	return &TCPAcceptor{tcpAddr: tcpAddr, options: options}, nil
}

// Listen listens and serve in the specified addr
func (this *TCPAcceptor) Listen(callback newSessionCallback) error {
	if callback == nil {
		return errors.New("dnet:Listen newSessionCallback is nil. ")
	}

	if !atomic.CompareAndSwapInt32(&this.started, 0, 1) {
		return errors.New("dnet:Listen acceptor is already started. ")
	}

	listener, err := net.ListenTCP("tcp", this.tcpAddr)
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
		tcpConn, err := NewTCPSession(conn.(*net.TCPConn), this.options...)
		if err != nil {
			return err
		}
		callback(tcpConn)
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
func DialTCP(address string, timeout time.Duration, options ...Option) (Session, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial(tcpAddr.Network(), address)
	if err != nil {
		return nil, err
	}

	return NewTCPSession(conn.(*net.TCPConn), options...)
}
