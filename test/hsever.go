package main

import (
	"fmt"
	"github.com/tagDong/dnet/http"
	"io/ioutil"
	http2 "net/http"
)

func test(w http2.ResponseWriter, r *http2.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Read failed:", err)
	}
	defer r.Body.Close()

	fmt.Println("http Post:", data, string(data))

	w.Write([]byte("OK"))
}

func main() {
	http.Register("/test", test)
	http.HttpStart("10.128.2.233:1234")
	fmt.Println("start 10.128.2.233:1234")
	select {}
}
