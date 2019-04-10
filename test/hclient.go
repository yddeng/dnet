package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func main() {
	url := "http://10.128.2.233:1234/test"
	contentType := "application/json;charset=utf-8"
	body := bytes.NewBuffer([]byte("sss"))
	resp, err := http.Post(url, contentType, body)
	if err != nil {
		fmt.Println("Report failed:", err)
		return
	}

	fmt.Println("read ", resp.Body)
}
