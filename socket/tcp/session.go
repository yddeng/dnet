package tcp

import (
	"errors"
	"github.com/yddeng/dnet"
	"log"
	"net"
	"sync/atomic"
	"time"
)

type Listener struct {
	listener *net.TCPListener
	started  int32
}

func NewListener(network, addr string) (*Listener, error) {
	tcpAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP(tcpAddr.Network(), tcpAddr)
	return &Listener{listener: listener}, err
}

func (l *Listener) Listen(newClient func(session dnet.Session)) error {
	if newClient == nil {
		return errors.New("newClient is nil")
	}

	if !atomic.CompareAndSwapInt32(&l.started, 0, 1) {
		return errors.New("tcpListener is started")
	}

	go func() {
		for {
			conn, err := l.listener.Accept()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					continue
				} else {
					log.Printf("listener err:%s", err)
					return
				}
			}
			newClient(newConn(conn))
		}
	}()

	return nil
}

func (l *Listener) Close() {
	if atomic.CompareAndSwapInt32(&l.started, 1, 0) {
		_ = l.listener.Close()
	}

}

func Dial(network, addr string, timeout time.Duration) (dnet.Session, error) {
	tcpAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial(tcpAddr.Network(), addr)
	if err != nil {
		return nil, err
	}

	return newConn(conn), nil
}
