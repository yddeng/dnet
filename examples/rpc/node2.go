package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/drpc"
	"github.com/yddeng/dnet/dtcp"
	"github.com/yddeng/dnet/examples/pb"
	"github.com/yddeng/dnet/examples/rpc/codec"
	"time"
)

func echo(replyer *drpc.Replyer, arg interface{}) {
	req := arg.(*pb.EchoToS)
	fmt.Println("echo", req.GetMsg())

	time.Sleep(time.Second * 9)
	err := replyer.Reply(&pb.EchoToC{Msg: proto.String(req.GetMsg())})
	if err != nil {
		fmt.Println(err)
	}
}

type channel struct {
	session dnet.Session
}

func (this *channel) SendRequest(req *drpc.Request) error {
	return this.session.Send(req)
}

func (this *channel) SendResponse(resp *drpc.Response) error {
	return this.session.Send(resp)
}

func main() {

	rpcServer := drpc.NewServer()
	rpcClient := drpc.NewClient()
	rpcServer.Register(proto.MessageName(&pb.EchoToS{}), echo)

	addr := "localhost:7756"
	session, err := dtcp.DialTCP("tcp", addr, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("conn ok,remote:%s\n", session.RemoteAddr())

	session.SetCodec(codec.NewRpcCodec())
	session.SetCloseCallBack(func(session dnet.Session, reason string) {
		fmt.Println("onClose", reason)
	})
	_ = session.Start(func(data interface{}, err2 error) {
		if err2 != nil {
			session.Close(err2.Error())
		} else {
			var err error
			switch data.(type) {
			case *drpc.Request:
				err = rpcServer.OnRPCRequest(&channel{session: session}, data.(*drpc.Request))
			case *drpc.Response:
				err = rpcClient.OnRPCResponse(data.(*drpc.Response))
			default:
				err = fmt.Errorf("invailed type")
			}
			if err != nil {
				fmt.Println("read", err)
			}
		}
	})

	msg := &pb.EchoToS{
		Msg: proto.String("hello node1,i'm node2"),
	}
	fmt.Println("Start Call")
	rpcClient.Call(&channel{session: session}, proto.MessageName(msg), msg, drpc.DefaultRPCTimeout, func(i interface{}, e error) {
		if e != nil {
			fmt.Println("Call", e)
			return
		}
		resp := i.(*pb.EchoToC)
		fmt.Println("node2 Call resp -->", resp.GetMsg())
	})

	select {}

}
