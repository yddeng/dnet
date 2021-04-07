## dnet

一个简单的tcp、websocket net的封装

```
type Session interface {
	// conn
    	NetConn() interface{}
    
    	// 获取远端地址
    	RemoteAddr() net.Addr
    
    	// 获取远端地址
    	LocalAddr() net.Addr
    
    	/*
    	 * 发送
    	 * 有编码器，任何数据都将经过编码器编码后发送
    	 * 没有编码器，仅能发送 []byte 类型数据。
    	 */
    	Send(o interface{}) error
    
    	// 给session绑定用户数据
    	SetContext(ctx interface{})
    
    	// 获取用户数据
    	Context() interface{}
    
    	/*
    	 先关闭读，待数据发送完毕再关闭连接
    	*/
    	Close(reason error)
    
    	// 是否已经关闭
    	IsClosed() bool
}
```

#### 编码(Codec)

自定义编解码器，实现如下接口：
```
//编解码器
type Codec interface {
	//编码
	Encode(interface{}) ([]byte, error)
	//解码
	Decode(reader io.Reader) (interface{}, error)
}
```

dnet默认的编码器，实现数据的沾包、分包。

#### example

echo

- protobuf 二进制协议
- handler 事件分发器
- codec 编解码器。消息 -- 格式: 消息头(消息len＋消息cmd+消息ID), 消息体

