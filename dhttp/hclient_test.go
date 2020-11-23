package dhttp

import (
	"fmt"
	"net/url"
	"testing"
)

func TestPostJson(t *testing.T) {
	req, err := PostJson("http://localhost:1234/echo", &Echo{Msg: "hello"})
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = req.Do()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(req.ToString())
}

func TestRequest_Param(t *testing.T) {
	req, err := NewRequest("http://localhost:1234/param", "POST")
	if err != nil {
		fmt.Println(err)
		return
	}

	data := make(url.Values)
	data.Set("1", "1")
	data.Set("1", "2")
	req.WriteParam(data)

	_, err = req.Do()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(req.ToString())
}

func TestRequest_PostFile(t *testing.T) {
	req, err := NewRequest("http://localhost:1234/upload", "POST")
	if err != nil {
		fmt.Println(err)
		return
	}

	req, err = req.PostFile("test.txt", "test/test.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := req.Do()
	fmt.Println(resp.StatusCode, err)
}
