package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/examples/cs/codec"
	"github.com/yddeng/dnet/examples/module/handler"
	"github.com/yddeng/dnet/examples/module/message"
	"github.com/yddeng/dnet/examples/pb"
	"time"
)

func echoToC(session dnet.Session, msg *message.Message) {
	data := msg.GetData().(*pb.EchoToS)
	fmt.Println("echo", data.GetMsg(), time.Now().String())

	if err := session.Send(message.NewMessage(0, &pb.EchoToC{Msg: proto.String("hello client")})); err != nil {
		fmt.Println(err)
	}
}

func main() {

	gHandler := handler.NewHandler()
	gHandler.RegisterCallBack(&pb.EchoToS{}, echoToC)

	addr := "localhost:1234"
	l, err := dnet.NewTCPAcceptor(addr,
		//dnet.WithTimeout(time.Second*5, 0), // 超时
		dnet.WithCodec(codec.NewCodec()),
		dnet.WithErrorCallback(func(session dnet.Session, err error) {
			fmt.Println("onError", err)
		}),
		dnet.WithMessageCallback(func(session dnet.Session, data interface{}) {
			gHandler.Dispatch(session, data.(*message.Message))
		}),
		dnet.WithCloseCallback(func(session dnet.Session, reason error) {
			fmt.Println("onClose", reason)
		}))
	if err != nil {
		fmt.Println(1, err)
		return
	}

	go func() {
		fmt.Println("server start on :", addr)
		if err = l.Listen(func(session dnet.Session) {
			fmt.Println("new client", session.RemoteAddr().String())
		}); err != nil {
			fmt.Println(2, err)
		}
	}()

	select {}
}
