package http

import "net/http"

var handler = http.NewServeMux()

func Register(pattern string, h func(http.ResponseWriter, *http.Request)) {
	handler.HandleFunc(pattern, h)
}

func HttpStart(addr string) {
	go func() {
		err := http.ListenAndServe(addr, handler)
		if err != nil {
			panic(err)
		}
	}()
}
