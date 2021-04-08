package dnet

import (
	"fmt"
	"net"
	"testing"
	"time"
)

type testTCPHandler struct{}

func (this *testTCPHandler) OnConnection(conn NetConn) {
	fmt.Println("new Conn", conn.RemoteAddr())
	session, err := NewTCPSession(conn,
		WithCloseCallback(func(session Session, reason error) {
			fmt.Println(session.RemoteAddr(), reason, "ss close")
		}),
		WithMessageCallback(func(session Session, message interface{}) {
			fmt.Println("ss", message)
		}),
		WithErrorCallback(func(session Session, err error) {
			fmt.Println("ss error", err)
		}))
	if err != nil {
		fmt.Println(err)
		return
	}
	time.Sleep(time.Millisecond * 200)
	fmt.Println(session.Send([]byte{4, 3, 2, 1}))
	fmt.Println(session.Send([]byte{4, 3, 2, 1}))
}

func TestNewTCPSession(t *testing.T) {
	go func() {
		ServeTCP(":4522", &testTCPHandler{})
	}()

	time.Sleep(time.Millisecond * 100)

	conn, err := DialTCP("127.0.0.1:4522", 0)
	if err != nil {
		fmt.Println("dialTcp", err)
		return
	}

	session, err := NewTCPSession(conn,
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
		fmt.Println("newTCPSession", err)
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

type testTCPHandler2 struct{}

func (this *testTCPHandler2) OnConnection(conn NetConn) {
	fmt.Println("new Conn", conn.RemoteAddr())

	go func() {
		buf := make([]byte, 1024)
		for {
			if n, err := conn.Read(buf); err != nil {
				fmt.Println(11, err)
				break
			} else {
				fmt.Println(buf[:n])
			}
		}
	}()

	go func() {
		time.Sleep(time.Millisecond * 500)
		conn.Write([]byte{4, 3, 2, 1})
	}()
}

func TestTCP(t *testing.T) {

	l, _ := net.Listen("tcp", ":4522")
	go func() {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go func() {
			buf := make([]byte, 1024)
			for {
				if n, err := c.(*net.TCPConn).Read(buf); err != nil {
					fmt.Println(11, err)
					break
				} else {
					fmt.Println(buf[:n])
					// 只读一条消息后 关闭读
					c.(*net.TCPConn).CloseRead()
				}
			}

		}()

		go func() {
			time.Sleep(time.Millisecond * 500)
			c.(*net.TCPConn).Write([]byte{4, 3, 2, 1})
		}()
	}()

	time.Sleep(time.Millisecond * 100)
	c, err := net.Dial("tcp", ":4522")
	if err != nil {
		fmt.Println(err)
	}

	go func() {

		buf := make([]byte, 1024)
		for {
			if n, err := c.(*net.TCPConn).Read(buf); err != nil {
				fmt.Println(22, err)
				break
			} else {
				fmt.Println(buf[:n])
			}
		}
	}()

	go func() {
		time.Sleep(time.Millisecond * 200)
		c.Write([]byte{1, 2, 3, 4})
		time.Sleep(time.Millisecond * 500)
		//c.(*net.TCPConn).CloseRead()
		n, err := c.Write([]byte{1, 2, 3, 4})
		fmt.Println(n, err)
	}()

	time.Sleep(time.Second * 2)
	c.Close()
	n, err := c.Write([]byte{1, 2, 3, 4})
	fmt.Println(33, n, err)
	time.Sleep(time.Second * 1)

}
