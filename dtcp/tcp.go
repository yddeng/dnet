package dtcp

import (
	"github.com/yddeng/dnet"
	"net"
	"sync/atomic"
	"time"
)

type TCPListener struct {
	listener *net.TCPListener
	started  int32
}

func NewTCPListener(network, addr string) (*TCPListener, error) {
	tcpAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP(tcpAddr.Network(), tcpAddr)
	return &TCPListener{listener: listener}, err
}

func (l *TCPListener) Listen(newClient func(session dnet.Session)) error {
	if newClient == nil {
		return dnet.ErrNewClientNil
	}

	if !atomic.CompareAndSwapInt32(&l.started, 0, 1) {
		return dnet.ErrStateFailed
	}

	go func() {
		for {
			conn, err := l.listener.Accept()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					continue
				} else {
					return
				}
			}
			newClient(NewTCPConn(conn.(*net.TCPConn)))
		}
	}()

	return nil
}

func (l *TCPListener) Addr() net.Addr {
	return l.listener.Addr()
}

func (l *TCPListener) Close() {
	if atomic.CompareAndSwapInt32(&l.started, 1, 0) {
		_ = l.listener.Close()
	}

}

func DialTCP(network, addr string, timeout time.Duration) (dnet.Session, error) {
	tcpAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial(tcpAddr.Network(), addr)
	if err != nil {
		return nil, err
	}

	return NewTCPConn(conn.(*net.TCPConn)), nil
}
