package dnet_test

import (
	"fmt"
	"github.com/tagDong/dnet/socket"
	"testing"
)

func TestStartTcpServe(t *testing.T) {
	socket.StartTcpServe("10.128.2.233:12345", func(session socket.StreamSession) {
		fmt.Println("newClient ", session.GetRemoteAddr())
		session.Start(func(data interface{}) {
			fmt.Println("Decode ", data)
		})

	})

	fmt.Println("------")

}
