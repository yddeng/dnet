package socket

import (
	"fmt"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/util"
	"testing"
	"time"
)

func TestNewTcpListener(t *testing.T) {
	addr := "localhost:1234"
	l, err := NewTcpListener("tcp", addr)
	if err != nil {
		fmt.Println(1, err)
		return
	}

	go func() {
		err := l.StartService(func(session dnet.Session) {
			fmt.Println("new client", session.RemoteAddr().String())
			session.SetCodec(util.NewCodec())
			//session.SetTimeout(3*time.Second, 0)
			session.SetCloseCallBack(func(reason string) {
				fmt.Println("session close", reason)
			})
			errr := session.Start(func(msg interface{}, err error) {
				if err != nil {
					session.Close(err.Error())
				} else {
					data := msg.([]byte)
					fmt.Println("server read", data)
					session.Send(data)
				}
			})
			if errr != nil {
				fmt.Println(2, err)
			}
		})
		if err != nil {
			fmt.Println(3, err)
		}
	}()

	//time.Sleep(10 * time.Second)
}

func TestTCPDial(t *testing.T) {
	addr := "localhost:1234"

	time.Sleep(time.Second)
	session, err := TCPDial("tcp", addr, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	session.SetCodec(util.NewCodec())
	session.SetCloseCallBack(func(reason string) {
		fmt.Println("session close", reason)
	})
	_ = session.Start(func(msg interface{}, err error) {
		if err != nil {
			session.Close(err.Error())
		} else {
			data := msg.([]byte)
			fmt.Println("client read", data)
		}
	})

	_ = session.Send([]byte{1, 2, 3})
	_ = session.Send([]byte{2, 2, 3, 4})
	_ = session.Send([]byte{3, 2, 3, 5, 6})
	fmt.Println(session.Send("code error"))
	time.Sleep(time.Second)
}

func TestNewWSListener(t *testing.T) {
	listener, err := NewWSListener("tcp", "127.0.0.1:1234", "/echo")
	if err != nil {
		fmt.Println(1, err)
		return
	}

	go func() {
		err = listener.StartService(func(session dnet.Session) {
			fmt.Println("new client", session.RemoteAddr().String())
			session.SetCloseCallBack(func(reason string) {
				fmt.Println("session close", reason)
			})
			errr := session.Start(func(msg interface{}, err error) {
				if err != nil {
					session.Close(err.Error())
				} else {
					data := msg.([]byte)
					fmt.Println("server read", data)
					session.Send(data)
				}
			})
			if errr != nil {
				fmt.Println(2, err)
			}
		})
		if err != nil {
			fmt.Println(3, err)
		}

	}()
}

func TestWSDial(t *testing.T) {
	addr := "localhost:1234"

	session, err := WSDial(addr, "/echo", 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	session.SetCloseCallBack(func(reason string) {
		fmt.Println("session close", reason)
	})
	_ = session.Start(func(msg interface{}, err error) {
		if err != nil {
			session.Close(err.Error())
		} else {
			data := msg.([]byte)
			fmt.Println("client read", data)
		}
	})

	_ = session.Send([]byte{1, 2, 3})
	_ = session.Send([]byte{2, 2, 3, 4})
	_ = session.Send([]byte{3, 2, 3, 5, 6})
	fmt.Println(session.Send("code error"))
	time.Sleep(time.Second)
}
