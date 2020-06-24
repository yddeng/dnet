package dhttp

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

type Echo struct {
	Msg string `json:"msg"`
}

func TestNewHttpServer(t *testing.T) {
	s := NewHttpServer("localhost:1234")

	s.HandleFuncJson("/echo", &Echo{}, func(w http.ResponseWriter, msg interface{}) {
		req := msg.(*Echo)
		fmt.Println(req.Msg)

		w.Write([]byte(req.Msg))
	})

	s.HandleFuncUrlParam("/param", func(w http.ResponseWriter, msg interface{}) {
		from := msg.(url.Values)
		fmt.Println(from)

		w.Write([]byte(from.Encode()))
	})

	s.Handle("/upload", HandleUpload)

	fmt.Println("listen")

	_ = s.Listen()

}
