package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/examples/cs/codec"
	"github.com/yddeng/dnet/examples/module/message"
	"github.com/yddeng/dnet/examples/pb"
	"time"
)

func main() {
	addr := "localhost:1234"
	conn, err := dnet.DialTCP(addr, 0)
	if err != nil {
		fmt.Println("dialTcp", err)
		return
	}

	session, err := dnet.NewTCPSession(conn,
		dnet.WithCodec(codec.NewCodec()),
		dnet.WithErrorCallback(func(session dnet.Session, err error) {
			fmt.Println("onError", err)
		}),
		dnet.WithMessageCallback(func(session dnet.Session, data interface{}) {
			fmt.Println("read ", data.(*message.Message).GetData())
		}),
		dnet.WithCloseCallback(func(session dnet.Session, reason error) {
			fmt.Println("onClose", reason)
		}))
	if err != nil {
		fmt.Println("newTCPSession", err)
		return
	}

	fmt.Printf("conn ok,remote:%s\n", session.RemoteAddr())

	fmt.Println(session.Send(message.NewMessage(0, &pb.EchoToS{Msg: proto.String("hi server")})))
	fmt.Println(session.Send(message.NewMessage(0, &pb.EchoToS{Msg: proto.String("hi server")})))
	fmt.Println(session.Send(message.NewMessage(0, &pb.EchoToS{Msg: proto.String("hi server")})))
	time.Sleep(5 * time.Second)

	select {}

}
