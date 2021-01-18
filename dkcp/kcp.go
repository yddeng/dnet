package dkcp

import (
	"github.com/xtaci/kcp-go"
	"github.com/yddeng/dnet"
	"net"
	"sync/atomic"
)

type KCPListener struct {
	listener *kcp.Listener
	started  int32
}

func NewKCPListener(laddr string) (*KCPListener, error) {
	listener, err := kcp.ListenWithOptions(laddr, nil, 0, 0)
	if err != nil {
		return nil, err
	}

	return &KCPListener{listener: listener}, err
}

func (l *KCPListener) Listen(newClient func(session dnet.Session)) error {
	if newClient == nil {
		return dnet.ErrNewClientNil
	}

	if !atomic.CompareAndSwapInt32(&l.started, 0, 1) {
		return dnet.ErrStateFailed
	}

	go func() {
		for {
			conn, err := l.listener.AcceptKCP()
			if err != nil {
				return
			}
			newClient(NewKCPConn(conn))
		}
	}()

	return nil
}

func (l *KCPListener) Addr() net.Addr {
	return l.listener.Addr()
}

func (l *KCPListener) Close() {
	if atomic.CompareAndSwapInt32(&l.started, 1, 0) {
		_ = l.listener.Close()
	}

}

func DialKCP(raddr string) (dnet.Session, error) {
	conn, err := kcp.DialWithOptions(raddr, nil, 0, 0)
	if err != nil {
		return nil, err
	}

	return NewKCPConn(conn), nil
}
