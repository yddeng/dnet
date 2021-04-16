package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/drpc"
	"github.com/yddeng/dnet/examples/pb"
	"github.com/yddeng/dnet/examples/rpc/codec"
	"sync/atomic"
	"time"
)

var times int32 = 0

func echo(replyer *drpc.Replier, arg interface{}) {
	req := arg.(*pb.EchoToS)
	fmt.Println("echo", req.GetMsg())

	t := atomic.AddInt32(&times, 1)
	if t == 1 {
		replyer.Reply(nil, errors.New("test rpc error"))
		return
	}

	replyer.Reply(&pb.EchoToC{Msg: proto.String("ok")}, nil)
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
	rpcServer.Register(proto.MessageName(&pb.EchoToS{}), echo)

	addr := "localhost:7756"
	go func() {
		if err := dnet.ServeTCP(addr, dnet.HandleFunc(func(conn dnet.NetConn) {
			fmt.Println("new client", conn.RemoteAddr().String())

			dnet.NewTCPSession(conn,
				dnet.WithCodec(codec.NewRpcCodec()),
				dnet.WithErrorCallback(func(session dnet.Session, err error) {
					fmt.Println("onError", err)
				}),
				dnet.WithCloseCallback(func(session dnet.Session, reason error) {
					fmt.Println("onClose", reason)
				}),
				dnet.WithMessageCallback(func(session dnet.Session, data interface{}) {
					var err error
					switch data.(type) {
					case *drpc.Request:
						err = rpcServer.OnRPCRequest(&channel{session: session}, data.(*drpc.Request))
					default:
						err = fmt.Errorf("invailed type")
					}
					if err != nil {
						fmt.Println("read", err)
					}
				}))
		})); err != nil {
			fmt.Println(err)
		}
	}()

	fmt.Println(addr)

	time.Sleep(time.Second * 20)
}
