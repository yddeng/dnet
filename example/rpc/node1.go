package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/example/pb"
	"github.com/yddeng/dnet/example/rpc/codec"
	"github.com/yddeng/dnet/rpc"
	"github.com/yddeng/dnet/socket"
	"reflect"
	"time"
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
	l, err := socket.NewTcpListener("tcp", addr)
	if err != nil {
		fmt.Println(1, err)
		return
	}

	err = l.StartService(func(session dnet.Session) {
		fmt.Println("new client", session.RemoteAddr().String())
		// 超时时间
		session.SetTimeout(10*time.Second, 0)
		session.SetCodec(codec.NewRpcCodec())
		session.SetCloseCallBack(func(reason string) {
			fmt.Println("onClose", reason)
		})

		fmt.Println("newClient ", session.RemoteAddr(), reflect.TypeOf(session.RemoteAddr()))
		errr := session.Start(func(data interface{}, err error) {
			//fmt.Println("data", data, "err", err)
			if err != nil {
				session.Close(err.Error())
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
		if errr != nil {
			fmt.Println(2, err)
		}

		time.Sleep(time.Second * 3)
		msg := &pb.EchoToS{
			Msg: proto.String("hello node2,i'm node1"),
		}
		fmt.Println("AsynCall")
		rpcClient.AsynCall(&channel{session: session}, msg, func(i interface{}, e error) {
			if e != nil {
				fmt.Println("AsynCall", e)
				return
			}
			resp := i.(*pb.EchoToC)
			fmt.Println("node1 AsynCall -->", resp.GetMsg())
		})

		fmt.Println("Post")
		rpcClient.Post(&channel{session: session}, msg)
	})
	if err != nil {
		fmt.Println(3, err)
	}

	fmt.Println(addr)

	select {}
}
