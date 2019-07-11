package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/tagDong/dnet/codec/protobuf"
	"github.com/tagDong/dnet/example/pb"
	"github.com/tagDong/dnet/socket"
	"github.com/tagDong/dnet/socket/tcp"
)

func main() {

	conn, err := tcp.NewTcpConnector("10.128.2.252:12345")
	if err != nil {
		panic(err)
	}
	fmt.Printf("conn ok,remote:%s ,local:%s\n", conn.RemoteAddr(), conn.LocalAddr())

	session := socket.NewSession(conn, protobuf.NewEncode(), protobuf.NewReader())
	session.Start(func(data interface{}) {
		fmt.Println("read ", data)
	})

	session.Send(&pb.EchoToS{Msg: proto.String("hi server")})

	select {}

}
