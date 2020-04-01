package dhttp

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

type Echo struct {
	Msg string `json:"msg"`
}

func TestNewHttpServer(t *testing.T) {
	s := NewHttpServer("localhost:1234")

	s.HandleFuncJson("/echo", &Echo{}, func(w http.ResponseWriter, msg interface{}) {
		req := msg.(*Echo)
		fmt.Println(req.Msg)
	})

	fmt.Println("listen")

	go func() {
		_ = s.Listen()
	}()

	time.Sleep(time.Second)
	resp, err := PostJson("http://localhost:1234/echo", &Echo{Msg: "hello"}, 0)
	fmt.Println(resp.StatusCode, err)
}
