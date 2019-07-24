package socket

import (
	"errors"
	"github.com/tagDong/dnet"
	"github.com/tagDong/dnet/codec"
	"github.com/tagDong/dnet/module/protocol"
	"github.com/tagDong/dnet/socket/tcp"
	"net"
)

type Server struct {
	protoc *protocol.Protocol
}

func NewServer(protoc *protocol.Protocol) *Server {
	return &Server{
		protoc: protoc,
	}
}

func (s *Server) StartTcpServe(addr string, newClient func(session dnet.Session)) error {
	if newClient == nil {
		return errors.New("newClient is nil")
	}

	listener, err := tcp.NewTCPListener(addr)
	if err != nil {
		return err
	}

	go s.tcpServe(listener, newClient)

	return nil
}

func (s *Server) tcpServe(listener *net.TCPListener, newClient func(dnet.Session)) {
	// 关闭监听
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				continue
			} else {
				return
			}
		}

		newClient(NewSession(conn, codec.NewCodec(s.protoc)))
	}

}
