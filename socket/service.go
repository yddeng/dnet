package socket

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/yddeng/dnet"
	tsocket "github.com/yddeng/dnet/socket/tcpsocket"
	wsocket "github.com/yddeng/dnet/socket/websocket"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

type TcpListener struct {
	listener *net.TCPListener
	started  int32
}

func NewTcpListener(network, addr string) (*TcpListener, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	case "":
		network = "tcp"
	default:
		return nil, fmt.Errorf("unknown network %s", network)
	}

	tcpAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP(network, tcpAddr)
	return &TcpListener{listener: listener}, err
}

func (l *TcpListener) StartService(newClient func(session dnet.Session)) error {
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
			newClient(tsocket.NewTcpSocket(conn))
		}
	}()

	return nil
}

func (l *TcpListener) Close() {
	if !atomic.CompareAndSwapInt32(&l.started, 1, 0) {
		_ = l.listener.Close()
	}

}

func TCPDial(network, addr string, timeout time.Duration) (dnet.Session, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	case "":
		network = "tcp"
	default:
		return nil, fmt.Errorf("unknown network %s", network)
	}

	_, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	return tsocket.NewTcpSocket(conn), nil
}

/*
 * websocket
 */

type WSListener struct {
	listener *net.TCPListener
	upgrader *websocket.Upgrader
	origin   string
	started  int32
}

func NewWSListener(network, addr, origin string, upgrader ...*websocket.Upgrader) (*WSListener, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	case "":
		network = "tcp"
	default:
		return nil, fmt.Errorf("unknown network %s", network)
	}

	tcpAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP(network, tcpAddr)
	if err != nil {
		return nil, err
	}

	l := &WSListener{
		listener: listener,
		origin:   origin,
	}

	if len(upgrader) > 0 {
		l.upgrader = upgrader[0]
	} else {
		l.upgrader = &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// allow all connections by default
				return true
			},
		}
	}
	return l, nil
}

func (this *WSListener) Close() {
	if !atomic.CompareAndSwapInt32(&this.started, 1, 0) {
		this.listener.Close()
	}
}

func (this *WSListener) StartService(newClient func(dnet.Session)) error {

	if newClient == nil {
		return errors.New("newClient is nil")
	}

	if !atomic.CompareAndSwapInt32(&this.started, 0, 1) {
		return errors.New("tcpListener is started")
	}

	http.HandleFunc(this.origin, func(w http.ResponseWriter, r *http.Request) {
		c, err := this.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("wssocket Upgrade failed:%s\n", err.Error())
			return
		}
		sess := wsocket.NewWebSocket(c)
		newClient(sess)
	})

	go func() {
		err := http.Serve(this.listener, nil)
		if err != nil {
			log.Printf("http.Serve() failed:%s\n", err.Error())
		}

		_ = this.listener.Close()
	}()

	return nil
}

func WSDial(addr, path string, timeout time.Duration) (dnet.Session, error) {
	u := url.URL{Scheme: "ws", Host: addr, Path: path}
	websocket.DefaultDialer.HandshakeTimeout = timeout
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return wsocket.NewWebSocket(conn), nil
}
