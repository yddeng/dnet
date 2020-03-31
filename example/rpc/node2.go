package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/example/pb"
	"github.com/yddeng/dnet/example/rpc/codec"
	"github.com/yddeng/dnet/rpc"
	"github.com/yddeng/dnet/socket"
	"github.com/yddeng/dnet/socket/tcp"
)

func echo(req *pb.EchoToS, resp *pb.EchoToC) {
	fmt.Println("echo", req.GetMsg())
	resp.Msg = proto.String(req.GetMsg())

}

type channel struct {
	session dnet.Session
}

func (this *channel) SendRequest(req *rpc.Request) error {
	return this.session.Send(req)
}

func (this *channel) SendResponse(resp *rpc.Response) error {
	return this.session.Send(resp)
}

func main() {

	rpcServer := rpc.NewServer()
	rpcClient := rpc.NewClient()
	rpcServer.Register(echo)

	addr := "localhost:7756"
	session, err := tcp.Dial("tcp", addr, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("conn ok,remote:%s\n", session.RemoteAddr())

	session.SetCodec(codec.NewRpcCodec())
	session.SetCloseCallBack(func(reason string) {
		fmt.Println("onClose", reason)
	})
	_ = session.Start(func(data interface{}, err2 error) {
		if err2 != nil {
			session.Close(err2.Error())
		} else {
			var err error
			switch data.(type) {
			case *rpc.Request:
				err = rpcServer.OnRPCRequest(&channel{session: session}, data.(*rpc.Request))
			case *rpc.Response:
				err = rpcClient.OnRPCResponse(data.(*rpc.Response))
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
	fmt.Println("AsynCall")
	rpcClient.AsynCall(&channel{session: session}, msg, func(i interface{}, e error) {
		if e != nil {
			fmt.Println("AsynCall", e)
			return
		}
		resp := i.(*pb.EchoToC)
		fmt.Println("node2 AsynCall -->", resp.GetMsg())
	})

	fmt.Println("Post")
	rpcClient.Post(&channel{session: session}, msg)

	fmt.Println("SynsCall")
	ret, err := rpcClient.SynsCall(&channel{session: session}, msg)
	if err != nil {
		fmt.Println("SynsCall", err)
		return
	}
	resp := ret.(*pb.EchoToC)
	fmt.Println("node2 SynsCall -->", resp.GetMsg())

	select {}

}
