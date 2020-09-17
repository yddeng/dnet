package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/drpc"
	"github.com/yddeng/dnet/dtcp"
	"github.com/yddeng/dnet/example/pb"
	"github.com/yddeng/dnet/example/rpc/codec"
	"time"
)

func echo(replyer *drpc.Replyer, arg interface{}) {
	req := arg.(*pb.EchoToS)
	fmt.Println("echo", req.GetMsg())

	replyer.Reply(&pb.EchoToC{Msg: proto.String(req.GetMsg())}, nil)
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
	l, err := dtcp.NewTCPListener("tcp", addr)
	if err != nil {
		fmt.Println(1, err)
		return
	}

	err = l.Listen(func(session dnet.Session) {
		fmt.Println("new client", session.RemoteAddr().String())
		// 超时时间
		session.SetTimeout(10*time.Second, 0)
		session.SetCodec(codec.NewRpcCodec())
		session.SetCloseCallBack(func(session dnet.Session, reason string) {
			fmt.Println("onClose", reason)
		})

		errr := session.Start(func(data interface{}, err error) {
			//fmt.Println("data", data, "err", err)
			if err != nil {
				session.Close(err.Error())
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
		if errr != nil {
			fmt.Println(2, err)
		}

		time.Sleep(time.Second * 3)
		msg := &pb.EchoToS{
			Msg: proto.String("hello node2,i'm node1"),
		}
		fmt.Println("Start AsynCall")
		rpcClient.AsynCall(&channel{session: session}, proto.MessageName(msg), msg, 8*time.Second, func(i interface{}, e error) {
			if e != nil {
				fmt.Println("AsynCall", e)
				return
			}
			resp := i.(*pb.EchoToC)
			fmt.Println("node1 AsynCall -->", resp.GetMsg())
		})

		fmt.Println("Start Post")
		rpcClient.Post(&channel{session: session}, proto.MessageName(msg), msg)
	})
	if err != nil {
		fmt.Println(3, err)
	}

	fmt.Println(addr)

	select {}
}
