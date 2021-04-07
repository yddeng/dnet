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

func NewTCPAcceptor(addr string, options ...Option) (*TCPAcceptor, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &TCPAcceptor{tcpAddr: tcpAddr, options: options}, nil
}

func (l *TCPAcceptor) Listen(callback newSessionCallback) error {
	if callback == nil {
		return errors.New("dnet:Listen newSessionCallback is nil. ")
	}

	if !atomic.CompareAndSwapInt32(&l.started, 0, 1) {
		return errors.New("dnet:Listen acceptor is already started. ")
	}

	listener, err := net.ListenTCP("tcp", l.tcpAddr)
	if err != nil {
		return err
	}
	l.listener = listener
	defer l.Stop()

	for {
		conn, err := l.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(time.Millisecond * 5)
				continue
			} else {
				return err
			}
		}
		tcpConn, err := NewTCPConn(conn.(*net.TCPConn), l.options...)
		if err != nil {
			return err
		}
		callback(tcpConn)
	}

}

func (l *TCPAcceptor) Addr() net.Addr {
	return l.listener.Addr()
}

func (l *TCPAcceptor) Stop() {
	if atomic.CompareAndSwapInt32(&l.started, 1, 0) {
		_ = l.listener.Close()
	}

}

func DialTCP(addr string, timeout time.Duration, options ...Option) (Session, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial(tcpAddr.Network(), addr)
	if err != nil {
		return nil, err
	}

	return NewTCPConn(conn.(*net.TCPConn), options...)
}
