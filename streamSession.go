package dnet

type StreamSession interface {
	// 获取远端地址
	GetRemoteAddr() string
	// 开启数据接收,返回实际数据，在上层反序列化
	StartReceive(func([]byte))
	//发送数据,在上层序列化后调用
	Send([]byte)
	// 给session绑定用户数据
	SetUserData(ud interface{})
	// 获取用户数据
	GetUserData() interface{}
	// 关闭连接
	Close() error
}
