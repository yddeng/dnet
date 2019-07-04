package socket

import (
	"errors"
	"github.com/tagDong/dnet/codec"
	"github.com/tagDong/dnet/socket/tcp"
	"net"
)

type Server struct {
	closeChan chan bool // 关服标识
	codec     codec.Codec
}

func newServer() *Server {
	return &Server{closeChan: make(chan bool)}
}

var server *Server

func StartTcpServe(addr string, newClient func(StreamSession)) error {
	if newClient == nil {
		return errors.New("newClient is nil")
	}

	listener, err := tcp.NewTCPListener(addr)
	if err != nil {
		return err
	}

	server = newServer()
	go server.tcpServe(listener, newClient)

	return nil
}

func (s *Server) Stop() {
	close(s.closeChan)
}

func (s *Server) tcpServe(listener *net.TCPListener, newClient func(StreamSession)) {
	// 关闭监听
	defer listener.Close()

	for {
		select {
		case <-s.closeChan:
			return
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				continue
			} else {
				return
			}
		}

		newClient(NewSession(conn, s.closeChan))
	}

}
