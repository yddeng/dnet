## dnet

一个简单的 `tcp`、`websocket` 的封装


### Session

```
type Session interface {
	// connection
	NetConn() interface{}
	
	// RemoteAddr returns the remote network address.
	RemoteAddr() net.Addr
	
	// LocalAddr returns the local network address.
	LocalAddr() net.Addr
	
	// Send data will be encoded by the encoder and sent
	Send(o interface{}) error
	
	// SetContext binding session data
	SetContext(ctx interface{})
	
	// Context returns binding session data
	Context() interface{}
	
	// Close closes the session.
	Close(reason error)
	
	// IsClosed returns has it been closed
	IsClosed() bool
}
```

### Functional options for session
```
// Options contains all options which will be applied when instantiating a session.
type Options struct {
	// when the send channel is full, BlockSend will block if it is true.
	// or it return queue full error code, if BlockSend is false.
	// it default is false
	BlockSend bool

	// capacity of the send channel. default net.defSendChannelSize
	SendChannelSize int

	// the deadline for read
	ReadTimeout time.Duration

	// the deadline for write
	WriteTimeout time.Duration

	// session will call the MsgCallback,if it has a message
	MsgCallback func(session Session, message interface{})

	// session will call the ErrorCallback,if it has a error
	ErrorCallback func(session Session, err error)

	// session will call the CloseCallback,if it is closed
	CloseCallback func(session Session, reason error)

	// encoder and decoder
	Codec Codec
}

// WithOptions accepts the whole options config.
func WithOptions(option *Options) Option {
	return func(opt *Options) {
		opt = option
	}
}

// WithBlockSend indicates whether it should block when the send channel full.
func WithBlockSend(bs bool) Option {
	return func(opt *Options) {
		opt.BlockSend = bs
	}
}

// WithSendChannelSize sets capacity of the send channel .
func WithSendChannelSize(size int) Option {
	return func(opt *Options) {
		opt.SendChannelSize = size
	}
}

// WithMessageCallback sets message callback.
func WithMessageCallback(msgCb func(session Session, message interface{})) Option {
	return func(opt *Options) {
		opt.MsgCallback = msgCb
	}
}

// WithErrorCallback sets error callback.
func WithErrorCallback(errCb func(session Session, err error)) Option {
	return func(opt *Options) {
		opt.ErrorCallback = errCb
	}
}

// WithTimeout sets the deadline of read/write.
func WithTimeout(readTimeout, writeTimeout time.Duration) Option {
	return func(opt *Options) {
		opt.ReadTimeout = readTimeout
		opt.WriteTimeout = writeTimeout
	}
}

// WithCodec sets codec.
func WithCodec(codec Codec) Option {
	return func(opt *Options) {
		opt.Codec = codec
	}
}

// WithCloseCallback sets close callback.
func WithCloseCallback(closeCallback func(session Session, reason error)) Option {
	return func(opt *Options) {
		opt.CloseCallback = closeCallback
	}
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

通过`WithCodec`设置会话的编解码器

`tcp`默认的编码器，实现数据的沾包、分包。

#### example

```
    // server  
    l, _ := NewTCPAcceptor(":4522",
        // 关闭连接回调
		WithCloseCallback(func(session Session, reason error) {
			fmt.Println(session.RemoteAddr(), reason, "ss close")
		}),
        // 消息回调
		WithMessageCallback(func(session Session, message interface{}) {
			fmt.Println("ss", message)
		}),
        // 错误回调
		WithErrorCallback(func(session Session, err error) {
			fmt.Println("ss error", err)
		}))
	defer l.Stop()

	go func() {
		_ = l.Listen(func(session Session) {
			time.Sleep(time.Millisecond * 200)
			fmt.Println(session.Send([]byte{4, 3, 2, 1}))
			fmt.Println(session.Send([]byte{4, 3, 2, 1}))
		})
	}()

    // client
    session, err := DialTCP("127.0.0.1:4522", 0,
		WithCloseCallback(func(session Session, reason error) {
			fmt.Println(session.RemoteAddr(), reason, "cc close")
		}),
		WithMessageCallback(func(session Session, message interface{}) {
			fmt.Println("cc", message)
		}),
		WithErrorCallback(func(session Session, err error) {
			fmt.Println("cc error", err)
		}))
	if err != nil {
		fmt.Println("dialTcp", err)
		return
	}

	fmt.Println(session.Send([]byte{1, 2, 3, 4}))
	fmt.Println(session.Send([]byte{1, 2, 3, 4, 5}))
	fmt.Println(session.Send([]byte{1, 2, 3, 4, 5, 6}))
	//fmt.Println(session.Send(123))

	time.Sleep(time.Second)
	fmt.Println(" ------- close ----------")
	session.Close(nil)
	fmt.Println(session.Send([]byte{1, 2, 3, 4}))
	time.Sleep(time.Second)

```

