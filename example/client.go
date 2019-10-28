package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/example/module/codec"
	"github.com/yddeng/dnet/example/module/message"
	"github.com/yddeng/dnet/example/pb"
	"time"
)

func main() {

	session, err := dnet.TCPDial("127.0.0.1:12345")
	if err != nil {
		panic(err)
	}
	fmt.Printf("conn ok,remote:%s\n", session.RemoteAddr())

	session.SetCodec(codec.NewCodec())
	session.SetCloseCallBack(func(reason string) {
		fmt.Println("onClose", reason)
	})
	_ = session.Start(func(data interface{}, err2 error) {
		//fmt.Println("data", data, "err", err)
		if err2 != nil {
			session.Close(err2.Error())
		} else {
			fmt.Println("read ", data.(*message.Message).GetData())
		}
	})

	fmt.Println(session.Send(message.NewMessage(0, &pb.EchoToS{Msg: proto.String("hi server")})))
	fmt.Println(session.Send(message.NewMessage(0, &pb.EchoToS{Msg: proto.String("hi server")})))
	time.Sleep(5 * time.Second)
	fmt.Println(session.Send(message.NewMessage(0, &pb.EchoToS{Msg: proto.String("hi server")})))

	select {}

}
