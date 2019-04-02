package socket

import (
	"fmt"
	"github.com/tagDong/dnet"
	"github.com/tagDong/dnet/socket/tcp"
	"net"
)

func StartTcpServe(addr string, newClient func(dnet.StreamSession)) {
	go func() {
		err := tcpServe(addr, newClient)
		if err != nil {
			panic(err)
		}
	}()
}

func tcpServe(addr string, newClient func(dnet.StreamSession)) error {
	if newClient == nil {
		return fmt.Errorf("newClient is nil")
	}

	listener, err := tcp.NewTCPListener(addr)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				continue
			} else {
				return err
			}
		}

		newClient(NewStreamSocket(conn))
	}
}
