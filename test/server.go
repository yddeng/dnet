package main

import (
	"fmt"
	"github.com/tagDong/dnet"
	"github.com/tagDong/dnet/socket"
)

func main() {
	socket.StartTcpServe("10.128.2.233:12345", func(session dnet.StreamSession) {
		fmt.Println("newClient ", session.GetRemoteAddr())
		session.StartReceive(func(bytes []byte) {
			fmt.Println("read ", bytes)
		})

		session.Send([]byte{1, 2, 3, 4})
		session.Send([]byte{4, 3, 2, 1})

	})

	fmt.Println("server start on : 10.128.2.233:12345")
	select {}
}
