package dnet

import (
	"time"
)

type Session interface {
	// 获取远端地址
	GetRemoteAddr() string
	// 开启数据接收
	Start(func(interface{}))
	//读写超时
	SetTimeout(readTimeout, writeTimeout time.Duration)
	//发送数据
	Send(data interface{})
	// 给session绑定用户数据
	SetUserData(ud interface{})
	// 获取用户数据
	GetUserData() interface{}
	// 关闭连接
	Close() error
}
