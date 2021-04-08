package dnet

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewTCPAcceptor(t *testing.T) {
	acceptor := NewTCPAcceptor(":4522")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		go func() {
			time.AfterFunc(time.Second*3, func() {
				acceptor.Stop()
				wg.Done()
			})
		}()

		err := acceptor.Serve(HandleFunc(func(conn NetConn) {
			buf := make([]byte, 8)
			n, err := conn.Read(buf)
			if err != nil {
				fmt.Println(err)
				conn.Close()
				return
			}
			fmt.Println(1111, buf[:n], n)
			time.Sleep(time.Millisecond * 300)
			conn.Write(buf[:n])

		}))
		if err != nil {
			fmt.Println(err)
		}
	}()

	time.Sleep(time.Millisecond * 500)
	conn, err := DialTCP(":4522", 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	n, err := conn.Write([]byte{1, 2, 3, 4, 5})
	fmt.Println(2222, n, err)

	buf := make([]byte, 8)
	n, err = conn.Read(buf)

	fmt.Println(3333, buf, n, err)
	time.Sleep(time.Millisecond * 500)
	conn.Close()

	wg.Wait()
	time.Sleep(time.Millisecond * 500)

}
