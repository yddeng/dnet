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
	fmt.Println("serve tcp", addr)
	if _, err := dnet.ServeTCP(addr, dnet.HandleFunc(func(conn dnet.NetConn) {
		fmt.Println("new client", conn.RemoteAddr().String())
		_ = dnet.NewTCPSession(conn,
			dnet.WithTimeout(time.Second*5, 0), // 超时
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
	})); err != nil {
		fmt.Println(err)
	}
}
