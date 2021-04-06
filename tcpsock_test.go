package dnet

import (
	"fmt"
	"testing"
	"time"
)

func TestTCP(t *testing.T) {
	l, _ := NewTCPListener("tcp", ":4522",
		WithCloseCallback(func(session Session, reason error) {
			fmt.Println(session.RemoteAddr(), reason, "ss close")
		}),
		WithCodec(NewCodec()),
		WithMessageCallback(func(message interface{}, err error) {
			fmt.Println("ss", message, err)
		}))
	go func() {
		_ = l.Listen(func(session Session) {
			time.Sleep(time.Millisecond * 500)
			fmt.Println(session.Send([]byte{4, 3, 2, 1}))
			fmt.Println(session.Send([]byte{4, 3, 2, 1}))
		})
	}()

	session, _ := DialTCP("tcp", "127.0.0.1:4522", 0,
		WithCloseCallback(func(session Session, reason error) {
			fmt.Println(session.RemoteAddr(), reason, "cc close")
		}),
		WithCodec(NewCodec()),
		WithMessageCallback(func(message interface{}, err error) {
			fmt.Println("cc", message, err)
		}))

	fmt.Println(session.Send(123))
	fmt.Println(session.Send([]byte{1, 2, 3, 4}))
	fmt.Println(session.Send([]byte{1, 2, 3, 4}))
	fmt.Println(session.Send([]byte{1, 2, 3, 4}))

	time.Sleep(time.Second)
	session.Close(nil)
	time.Sleep(time.Second)

}
