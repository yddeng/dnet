package dnet

import (
	"fmt"
	"testing"
	"time"
)

func TestNewWSSession(t *testing.T) {
	l, _ := NewWSAcceptor(":4522",
		WithCloseCallback(func(session Session, reason error) {
			fmt.Println(session.RemoteAddr(), reason, "ss close")
		}),
		WithMessageCallback(func(session Session, message interface{}) {
			fmt.Println("ss", message)
		}),
		WithErrorCallback(func(session Session, err error) {
			fmt.Println("ss error", err)
		}))
	defer l.Stop()

	go func() {
		_ = l.Listen(func(session Session) {
			time.Sleep(time.Millisecond * 200)
			fmt.Println(session.Send([]byte{4, 3, 2, 1}))
			fmt.Println(session.Send([]byte{4, 3, 2, 1}))
		})
	}()

	time.Sleep(time.Millisecond * 100)
	session, err := DialWS("127.0.0.1:4522", 0,
		WithCloseCallback(func(session Session, reason error) {
			fmt.Println(session.RemoteAddr(), reason, "cc close")
		}),
		WithMessageCallback(func(session Session, message interface{}) {
			fmt.Println("cc", message)
		}),
		WithErrorCallback(func(session Session, err error) {
			fmt.Println("cc error", err)
		}))
	if err != nil {
		fmt.Println("dialWs", err)
		return
	}

	fmt.Println(session.Send([]byte{1, 2, 3, 4}))
	fmt.Println(session.Send([]byte{1, 2, 3, 4, 5}))
	fmt.Println(session.Send([]byte{1, 2, 3, 4, 5, 6}))
	//fmt.Println(session.Send(123))

	time.Sleep(time.Second)
	fmt.Println(" ------- close ----------")
	session.Close(nil)
	fmt.Println(session.Send([]byte{1, 2, 3, 4}))
	time.Sleep(time.Second)
}
