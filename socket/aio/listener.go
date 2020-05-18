package aio

import (
	"errors"
	"fmt"
	"github.com/yddeng/dnet"
	"log"
	"net"
	"reflect"
	"sync/atomic"
)

type AioListener struct {
	listener net.Listener
	service  *Service
	started  int32
}

func NewListener(network, addr string, service *Service) (*AioListener, error) {
	tcpAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP(tcpAddr.Network(), tcpAddr)
	if err != nil {
		return nil, err
	}

	return &AioListener{
		listener: listener,
		service:  service,
	}, nil
}

func (l *AioListener) Listen(newClient func(session dnet.Session)) error {
	if newClient == nil {
		return errors.New("newClient is nil")
	}

	if !atomic.CompareAndSwapInt32(&l.started, 0, 1) {
		return errors.New("tcpListener is started")
	}

	go l.service.Start()

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
			fmt.Println("new", conn.RemoteAddr())

			newClient(newAioConn(socketFD(conn), conn, l.service.NextEventLoop()))
		}
	}()

	return nil
}

func (l *AioListener) Close() {
	if atomic.CompareAndSwapInt32(&l.started, 1, 0) {
		_ = l.listener.Close()
		l.service.Stop()
	}

}

func socketFD(conn net.Conn) int {
	//tls := reflect.TypeOf(conn.UnderlyingConn()) == reflect.TypeOf(&tls.Conn{})
	// Extract the file descriptor associated with the connection
	//connVal := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn").Elem()
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	//if tls {
	//	tcpConn = reflect.Indirect(tcpConn.Elem())
	//}
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")

	return int(pfdVal.FieldByName("Sysfd").Int())
}

func Dial(addr string, loop *EventLoop) (dnet.Session, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{Timeout: 0}
	conn, err := dialer.Dial(tcpAddr.Network(), addr)
	if err != nil {
		return nil, err
	}

	return newAioConn(socketFD(conn), conn, loop), nil
}
