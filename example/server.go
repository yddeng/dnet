package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/tagDong/dnet"
	"github.com/tagDong/dnet/example/pb"
	handler2 "github.com/tagDong/dnet/handler"
	"github.com/tagDong/dnet/socket"
)

func echoToC(session dnet.Session, msg proto.Message) {
	data := msg.(*pb.EchoToS)
	fmt.Println(data.GetMsg())

	session.Send(&pb.EchoToC{Msg: proto.String("hello client")})
}

func main() {

	handler := handler2.NewHandler()
	handler.Register(&pb.EchoToS{}, echoToC)

	server := socket.NewServer()
	server.StartTcpServe("10.128.2.252:12345", func(session dnet.Session) {
		fmt.Println("newClient ", session.GetRemoteAddr())
		session.Start(func(data interface{}) {
			handler.Dispatch(session, data.(proto.Message))
		})

	})

	fmt.Println("server start on : 10.128.2.233:12345")
	select {}
}
