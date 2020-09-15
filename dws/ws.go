package dws

import (
	"github.com/gorilla/websocket"
	"github.com/yddeng/dnet"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

type WSListener struct {
	listener *net.TCPListener
	upgrader *websocket.Upgrader
	origin   string
	started  int32
}

func NewWSListener(network, addr, origin string, upgrader ...*websocket.Upgrader) (*WSListener, error) {
	tcpAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP(tcpAddr.Network(), tcpAddr)
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

func (this *WSListener) Listen(newClient func(dnet.Session)) error {

	if newClient == nil {
		return dnet.ErrNewClientNil
	}

	if !atomic.CompareAndSwapInt32(&this.started, 0, 1) {
		return dnet.ErrStateFailed
	}

	http.HandleFunc(this.origin, func(w http.ResponseWriter, r *http.Request) {
		c, err := this.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("wssocket Upgrade failed:%s\n", err.Error())
			return
		}
		newClient(NewWSConn(c))
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

func DialWS(addr, path string, timeout time.Duration) (dnet.Session, error) {
	u := url.URL{Scheme: "ws", Host: addr, Path: path}
	websocket.DefaultDialer.HandshakeTimeout = timeout
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return NewWSConn(conn), nil
}
