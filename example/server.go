package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/example/module/codec"
	"github.com/yddeng/dnet/example/module/handler"
	"github.com/yddeng/dnet/example/module/message"
	"github.com/yddeng/dnet/example/pb"
	"reflect"
	"time"
)

func echoToC(session dnet.Session, msg *message.Message) {
	data := msg.GetData().(*pb.EchoToS)
	fmt.Println("echo", data.GetMsg())

	_ = session.Send(message.NewMessage(0, &pb.EchoToC{Msg: proto.String("hello client")}))
}

func main() {

	gHandler := handler.NewHandler()
	gHandler.RegisterCallBack(&pb.EchoToS{}, echoToC)

	_ = dnet.StartTcpServe("127.0.0.1:12345", func(session dnet.Session) {
		// 超时时间
		session.SetTimeout(10*time.Second, 0)
		session.SetCodec(codec.NewCodec())
		session.SetCloseCallBack(func(reason string) {
			fmt.Println("onClose", reason)
		})
		fmt.Println("newClient ", session.RemoteAddr(), reflect.TypeOf(session.RemoteAddr()))
		_ = session.Start(func(data interface{}, err error) {
			//fmt.Println("data", data, "err", err)
			if err != nil {
				session.Close(err.Error())
			} else {
				gHandler.Dispatch(session, data.(*message.Message))
			}
		})

	})

	fmt.Println("server start on : 10.128.2.233:12345")
	select {}
}
