package dnet

import (
	"errors"
	"github.com/yddeng/dnet/net/tcp"
	"net"
)

func StartTcpServe(addr string, newClient func(session Session)) error {
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

func tcpServe(listener *net.TCPListener, newClient func(Session)) {
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
		//fmt.Println(conn, reflect.TypeOf(conn))

		newClient(NewSocket(conn))
	}

}

func TCPDial(addr string) (Session, error) {
	conn, err := tcp.NewTcpConnector(addr)
	if err != nil {
		return nil, err
	}

	return NewSocket(conn), nil
}
