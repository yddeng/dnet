package dnet

import (
	"fmt"
	"testing"
	"time"
)

func TestNewWSSession(t *testing.T) {
	go func() {
		ServeWSFunc(":4522", func(conn NetConn) {
			fmt.Println("new Conn", conn.RemoteAddr())
			session := NewWSSession(conn,
				WithCloseCallback(func(session Session, reason error) {
					fmt.Println(session.RemoteAddr(), reason, "ss close")
				}),
				WithMessageCallback(func(session Session, message interface{}) {
					fmt.Println("ss", message)
				}),
				WithErrorCallback(func(session Session, err error) {
					fmt.Println("ss error", err)
				}))
			time.Sleep(time.Millisecond * 200)
			fmt.Println(session.Send([]byte{4, 3, 2, 1}))
			fmt.Println(session.Send([]byte{4, 3, 2, 1}))
		})
	}()

	//http.HandleFunc()

	time.Sleep(time.Millisecond * 100)
	wsConn, err := DialWS("127.0.0.1:4522", 0)
	if err != nil {
		fmt.Println("dialWs", err)
		return
	}

	session := NewWSSession(wsConn,
		WithCloseCallback(func(session Session, reason error) {
			fmt.Println(session.RemoteAddr(), reason, "cc close")
		}),
		WithMessageCallback(func(session Session, message interface{}) {
			fmt.Println("cc", message)
		}),
		WithErrorCallback(func(session Session, err error) {
			fmt.Println("cc error", err)
		}))

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
