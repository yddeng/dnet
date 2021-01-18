package dkcp

import (
	"fmt"
	"github.com/yddeng/dnet"
	"io"
	"testing"
	"time"
)

type codec struct {
	buf []byte
}

func newCodec() *codec {
	return &codec{buf: make([]byte, 4096)}
}

func (_ *codec) Encode(o interface{}) ([]byte, error) {
	return o.([]byte), nil
}

func (this *codec) Decode(reader io.Reader) (interface{}, error) {
	n, err := reader.Read(this.buf)
	if err != nil {
		return nil, err
	}
	return this.buf[:n], nil
}

func server(laddr string) {
	l, err := NewKCPListener(laddr)
	if err != nil {
		panic(err)
	}

	l.Listen(func(session dnet.Session) {
		session.SetCodec(newCodec())
		session.SetCloseCallBack(func(session dnet.Session, reason string) {
			fmt.Println(session.RemoteAddr(), "close", reason)
		})
		session.Start(func(message interface{}, err error) {
			if err != nil {
				fmt.Println(err)
				session.Close(err.Error())
			} else {
				fmt.Println("server read", message)
				session.Send(message)
			}
		})
	})
}

func client(raddr string) {
	session, err := DialKCP(raddr)
	if err != nil {
		panic(err)
	}

	session.SetCodec(newCodec())
	session.SetCloseCallBack(func(session dnet.Session, reason string) {
		fmt.Println(session.RemoteAddr(), "close", reason)
	})
	session.Start(func(message interface{}, err error) {
		if err != nil {
			fmt.Println(err)
			session.Close(err.Error())
		} else {
			fmt.Println("client read", message)
		}
	})

	time.Sleep(time.Second)
	fmt.Println(session.Send([]byte("hello")))

}

func TestKCP(t *testing.T) {
	addr := "127.0.0.1:12345"
	server(addr)
	client(addr)

	select {}
}
