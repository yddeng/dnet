package socket

import (
	"errors"
	"github.com/tagDong/dnet"
	"github.com/tagDong/dnet/socket/tcp"
	"net"
)

func StartTcpServe(addr string, newClient func(session dnet.Session)) error {
	if newClient == nil {
		return errors.New("newClient is nil")
	}

	listener, err := tcp.NewTCPListener(addr)
	if err != nil {
		return err
	}

	go tcpServe(listener, newClient)

	return nil
}

func tcpServe(listener *net.TCPListener, newClient func(dnet.Session)) {
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

		newClient(NewSession(conn))
	}

}

func SessionConnector(addr string) (dnet.Session, error) {
	conn, err := tcp.NewTcpConnector(addr)
	if err != nil {
		return nil, err
	}

	return NewSession(conn), nil
}
