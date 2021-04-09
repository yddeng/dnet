package drpc

import (
	"fmt"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/examples/rpc/codec"
	"testing"
	"time"
)

type EchoToS struct {
	Msg string `json:"msg,omitempty"`
}

type EchoToC struct {
	Msg string `json:"msg,omitempty"`
}

func echo(replyer *Replier, arg interface{}) {
	req := arg.(*EchoToS)
	fmt.Println("echo", req.Msg)
	replyer.Reply(&EchoToC{Msg: req.Msg})
}

type channel struct {
	session dnet.Session
}

func (this *channel) SendRequest(req *Request) error {
	return this.session.Send(req)
}

func (this *channel) SendResponse(resp *Response) error {
	return this.session.Send(resp)
}

func TestNewServer(t *testing.T) {

	rpcServer := NewServer()
	rpcClient := NewClient()
	rpcServer.Register("echo", echo)

	addr := "localhost:7756"
	go func() {
		if err := dnet.ServeTCP(addr, dnet.HandleFunc(func(conn dnet.NetConn) {
			fmt.Println("new client", conn.RemoteAddr().String())

			session := dnet.NewTCPSession(conn,
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
					case *Request:
						err = rpcServer.OnRPCRequest(&channel{session: session}, data.(*Request))
					case *Response:
						err = rpcClient.OnRPCResponse(data.(*Response))
					default:
						err = fmt.Errorf("invailed type")
					}
					if err != nil {
						fmt.Println("read", err)
					}
				}))

			time.Sleep(time.Second * 3)
			msg := &EchoToS{
				Msg: "hello node2,i'm node1",
			}
			fmt.Println("Start Call")
			rpcClient.Call(&channel{session: session}, "echo", msg, DefaultRPCTimeout, func(i interface{}, e error) {
				if e != nil {
					fmt.Println("Call", e)
					return
				}
				resp := i.(*EchoToC)
				fmt.Println("node1 Call resp -->", resp.Msg)
			})

		})); err != nil {
			fmt.Println(err)
		}
	}()

	fmt.Println(addr)

	time.Sleep(time.Second * 20)
}
