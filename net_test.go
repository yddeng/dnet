package dnet_test

import (
	"fmt"
	"github.com/yddeng/dnet"
	"testing"
)

func TestStartTcpServe(t *testing.T) {
	addr := "localhost:1234"
	_ = dnet.StartTcpServe(addr, func(session dnet.Session) {
		fmt.Println("new client", session.RemoteAddr().String())
		session.SetCodec(dnet.NewDefCodec())
		//session.SetTimeout(3*time.Second, 0)
		session.SetCloseCallBack(func(reason string) {
			fmt.Println("session close", reason)
		})
		_ = session.Start(func(msg interface{}, err error) {
			if err != nil {
				session.Close(err.Error())
			} else {
				data := msg.([]byte)
				fmt.Println("server read", data)
				session.Send(data)
			}
		})
	})

}

func TestTCPDial(t *testing.T) {
	addr := "localhost:1234"

	session, _ := dnet.TCPDial(addr)
	session.SetCodec(dnet.NewDefCodec())
	session.SetCloseCallBack(func(reason string) {
		fmt.Println("session close", reason)
	})
	_ = session.Start(func(msg interface{}, err error) {
		if err != nil {
			session.Close(err.Error())
		} else {
			data := msg.([]byte)
			fmt.Println("client read", data)
		}
	})

	_ = session.Send([]byte{1, 2, 3})
	_ = session.Send([]byte{2, 2, 3, 4})
	_ = session.Send([]byte{3, 2, 3, 5, 6})
	fmt.Println(session.Send("teststring"))
}
