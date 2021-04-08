package dnet

import (
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

type WSAcceptor struct {
	tcpAddr  *net.TCPAddr
	handler  *wsHandler
	listener *net.TCPListener
	started  int32
}

// NewWSAcceptor returns a new instance of WSAcceptor
func NewWSAcceptor(address string, options ...Option) (*WSAcceptor, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	return &WSAcceptor{
		tcpAddr: tcpAddr,
		handler: &wsHandler{
			options: options,
			upgrader: &websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					// allow all connections by default
					return true
				},
			},
		},
	}, nil
}

type wsHandler struct {
	upgrader *websocket.Upgrader
	options  []Option
	callback newSessionCallback
}

func (h *wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("dnet:ServeHTTP WSSession Upgrade failed, %s\n", err.Error())
		return
	}
	wsSession, err := NewWSSession(c, h.options...)
	if err != nil {
		log.Printf("dnet:ServeHTTP NewWSSession failed, %s\n", err.Error())
		return
	}
	h.callback(wsSession)
}

// Listen listens and serve in the specified addr
func (this *WSAcceptor) Listen(callback newSessionCallback) error {
	if callback == nil {
		return errors.New("dnet:WSAcceptor Listen newSessionCallback is nil. ")
	}
	this.handler.callback = callback

	if !atomic.CompareAndSwapInt32(&this.started, 0, 1) {
		return errors.New("dnet:WSAcceptor Listen acceptor is already started. ")
	}

	listener, err := net.ListenTCP("tcp", this.tcpAddr)
	if err != nil {
		return errors.New("dnet:WSAcceptor ListenTCP failed, " + err.Error())
	}
	this.listener = listener
	defer this.Stop()

	if err = http.Serve(this.listener, this.handler); err != nil {
		log.Printf("dnet:WSAcceptor Serve failed, %s\n", err.Error())
	}

	return nil
}

// Addr returns the addr the acceptor will listen on
func (this *WSAcceptor) Addr() net.Addr {
	return this.listener.Addr()
}

// Stop stops the acceptor
func (this *WSAcceptor) Stop() {
	if !atomic.CompareAndSwapInt32(&this.started, 1, 0) {
		_ = this.listener.Close()
	}
}

func DialWS(host string, timeout time.Duration, options ...Option) (Session, error) {
	u := url.URL{Scheme: "ws", Host: host}
	websocket.DefaultDialer.HandshakeTimeout = timeout
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return NewWSSession(conn, options...)
}
