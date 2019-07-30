package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/tagDong/dnet"
	"github.com/tagDong/dnet/codec"
	"github.com/tagDong/dnet/example/pb"
	"github.com/tagDong/dnet/module/message"
	"github.com/tagDong/dnet/socket"
	"github.com/tagDong/dnet/socket/tcp"
)

func main() {

	conn, err := tcp.NewTcpConnector("10.128.2.252:12345")
	if err != nil {
		panic(err)
	}
	fmt.Printf("conn ok,remote:%s ,local:%s\n", conn.RemoteAddr(), conn.LocalAddr())

	session := socket.NewSession(conn)
	session.SetCodec(codec.NewCodec())
	session.Start(func(data interface{}) {
		fmt.Println("read ", data.(dnet.Message).GetData())
		//session.Send(message.NewMessage(0, &pb.EchoToS{Msg: proto.String("hi server 1")}))
	})

	session.Send(message.NewMessage(0, &pb.EchoToS{Msg: proto.String("hi server")}))
	session.Send(message.NewMessage(0, &pb.EchoToS{Msg: proto.String("hi server")}))
	session.Send(message.NewMessage(0, &pb.EchoToS{Msg: proto.String("hi server")}))

	select {}

}
