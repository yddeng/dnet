package socket

import (
	"time"
)

type StreamSession interface {
	// 获取远端地址
	GetRemoteAddr() string
	// 开启数据接收,返回完整的数据，由上层进行解码
	Start(func(interface{}))
	//读写超时
	SetTimeout(readTimeout, writeTimeout time.Duration)
	//发送数据,在上层序列化后调用
	Send([]byte)
	// 给session绑定用户数据
	SetUserData(ud interface{})
	// 获取用户数据
	GetUserData() interface{}
	// 关闭连接
	Close() error
}
