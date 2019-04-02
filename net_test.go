package dnet_test

import (
	"fmt"
	"github.com/tagDong/dnet"
	"github.com/tagDong/dnet/socket"
	"testing"
)

func TestStartTcpServe(t *testing.T) {
	socket.StartTcpServe("10.128.2.233:12345", func(session dnet.StreamSession) {
		fmt.Println("newClient ", session.GetRemoteAddr())
		session.StartReceive(func(bytes []byte) {
			fmt.Println("Decode ", bytes)
		})

	})

	fmt.Println("------")
}
