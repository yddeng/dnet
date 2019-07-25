package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/tagDong/dnet"
	"github.com/tagDong/dnet/example/pb"
	"github.com/tagDong/dnet/module/handler"
	"github.com/tagDong/dnet/module/message"
	"github.com/tagDong/dnet/socket"
	"time"
)

func echoToC(session dnet.Session, msg dnet.Message) {
	data := msg.GetData().(*pb.EchoToS)
	fmt.Println("echo", data.GetMsg())

	session.Send(message.NewMessage(0, &pb.EchoToC{Msg: proto.String("hello client")}))
}

func main() {

	gHandler := handler.NewHandler()
	gHandler.RegisterCallBack(&pb.EchoToS{}, echoToC)

	socket.StartTcpServe("10.128.2.252:12345", func(session dnet.Session) {
		session.SetTimeout(8*time.Second, 0)
		fmt.Println("newClient ", session.GetRemoteAddr())
		session.Start(func(data interface{}) {
			gHandler.Dispatch(session, data.(dnet.Message))
		})

	})

	fmt.Println("server start on : 10.128.2.233:12345")
	select {}
}
