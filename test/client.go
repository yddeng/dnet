package main

import (
	"fmt"
	"github.com/tagDong/dnet/socket"
	"github.com/tagDong/dnet/socket/tcp"
)

func main() {
	conn, err := tcp.NewTcpConnector("10.128.2.252:12345")
	if err != nil {
		panic(err)
	}
	fmt.Printf("conn ok,remote:%s ,local:%s\n", conn.RemoteAddr(), conn.LocalAddr())
	closeChan := make(chan bool)

	session := socket.NewSession(conn, closeChan)
	session.Start(func(data interface{}) {
		fmt.Println("read ", data)
	})
	session.Send([]byte{0, 4, 0, 1})
	session.Send([]byte{0, 2, 0, 3, 0, 3})

	select {}

}
