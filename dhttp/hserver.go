package dhttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

type HandlerFunc func(w http.ResponseWriter, msg interface{}) // 回调方法

type HttpServer struct {
	handlers   *http.ServeMux
	listenAddr string
}

func NewHttpServer(addr string) *HttpServer {
	s := new(HttpServer)
	s.handlers = http.NewServeMux()
	s.listenAddr = addr

	return s
}

// post
// 注册json请求,将请求数据通过json转成对应的结构
// 路由，结构，方法
func (s *HttpServer) HandleFuncJson(route string, elem interface{}, fn HandlerFunc) {
	elemT := reflect.TypeOf(elem)

	s.handlers.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		msg := reflect.New(elemT.Elem()).Interface()
		err := json.NewDecoder(r.Body).Decode(&msg)
		defer r.Body.Close()

		if err != nil {
			serveError(w, 404, err.Error())
			return
		}

		fn(w, msg)
	})
}

// get
// 解析url的地址参数，如果参数不够
// 路由，参数，方法
func (s *HttpServer) HandleFuncUrlParam(route string, fn HandlerFunc) {
	s.handlers.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		fn(w, r.Form)
	})
}

// 注册http默认方法
func (s *HttpServer) Handle(route string, fn http.Handler) {
	s.handlers.Handle(route, fn)
}

func (s *HttpServer) Listen() error {
	return http.ListenAndServe(s.listenAddr, s.handlers)
}

func httpHeader(w *http.ResponseWriter) {
	//跨域
	(*w).Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	(*w).Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	(*w).Header().Set("content-type", "application/json")             //返回数据格式是json
}

func serveError(w http.ResponseWriter, status int, txt string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintln(w, txt)
}
