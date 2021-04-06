package dnet

import (
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type TCPConn struct {
	opts *Options

	state byte
	conn  *net.TCPConn

	context interface{} // 用户数据

	sendNotifyCh  chan struct{} // 发送消息通知
	sendMessageCh chan *message // 发送队列

	closeReason error

	lock sync.Mutex
}

type message struct {
	needEncode bool
	data       interface{}
}

func NewTCPConn(conn *net.TCPConn, options ...Option) (*TCPConn, error) {
	op := loadOptions(options...)
	if op.SendChannelSize <= 0 {
		op.SendChannelSize = sendBufChanSize
	}
	if op.ReadBufferSize <= 0 {
		op.ReadBufferSize = readBufferSize
	}
	if op.MsgCallback == nil {
		return nil, ErrNilMsgCallBack
	}

	tcpConn := &TCPConn{
		opts:          op,
		conn:          conn,
		state:         started,
		sendNotifyCh:  make(chan struct{}, 1),
		sendMessageCh: make(chan *message, op.SendChannelSize),
	}

	go tcpConn.readThread()
	go tcpConn.writeThread()

	return tcpConn, nil
}

func (this *TCPConn) NetConn() interface{} {
	return this.conn
}

func (this *TCPConn) LocalAddr() net.Addr {
	return this.conn.LocalAddr()
}

//对端地址
func (this *TCPConn) RemoteAddr() net.Addr {
	return this.conn.RemoteAddr()
}

func (this *TCPConn) SetContext(context interface{}) {
	this.lock.Lock()
	this.context = context
	this.lock.Unlock()
}

func (this *TCPConn) Context() interface{} {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.context
}

func (this *TCPConn) isClose() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.state == closed
}

// 接收线程
func (this *TCPConn) readThread() {
	var buffer = make([]byte, this.opts.ReadBufferSize)
	for {
		if this.isClose() {
			return
		}

		if this.opts.ReadTimeout > 0 {
			_ = this.conn.SetReadDeadline(time.Now().Add(this.opts.ReadTimeout))
		}

		if n, err := this.conn.Read(buffer); err != nil {
			this.opts.MsgCallback(nil, err)
			if err == io.EOF {
				break
			}
		} else if this.opts.Decoder != nil {
			if msg, err := this.opts.Decoder.Decode(buffer[:n]); err != nil {
				this.opts.MsgCallback(nil, err)
			} else if msg != nil {
				this.opts.MsgCallback(msg, nil)
			}
		} else {
			this.opts.MsgCallback(buffer[:n], nil)
		}
	}
}

// 发送线程
// 关闭连接时，发送完后再关闭
func (this *TCPConn) writeThread() {
	for {
		select {
		case msg := <-this.sendMessageCh:
			var err error
			var data []byte

			if msg.needEncode {
				// 需要编码的消息
				data, err = this.opts.Encoder.Encode(msg.data)
				if err != nil {
					this.opts.MsgCallback(nil, err)
					break
				}
			} else {
				data, _ = msg.data.([]byte)
			}

			if data != nil && len(data) != 0 {
				// 发送的消息
				if this.opts.WriteTimeout > 0 {
					_ = this.conn.SetWriteDeadline(time.Now().Add(this.opts.WriteTimeout))
				}

				_, err = this.conn.Write(data)
				if err != nil {
					this.opts.MsgCallback(nil, err)
				}
			}

		default:
			closed := this.isClose()
			if closed {
				this.close()
				return
			} else {
				// 等待发送事件
				<-this.sendNotifyCh
			}
		}

	}
}

func (this *TCPConn) Send(o interface{}) error {
	if o == nil {
		return ErrSendMsgNil
	}

	this.lock.Lock()
	if this.state != started {
		this.lock.Unlock()
		return ErrStateFailed
	}
	this.lock.Unlock()

	if this.opts.Encoder == nil {
		if _, ok := o.([]byte); !ok {
			return ErrSendTypeFailed
		}
	}

	if !this.opts.BlockSend {
		if len(this.sendMessageCh) == this.opts.SendChannelSize {
			return ErrSendChanFull
		}
	}

	this.sendMessageCh <- &message{
		needEncode: this.opts.Encoder != nil,
		data:       o,
	}
	SendNotifyChan(this.sendNotifyCh)
	return nil
}

/*
 主动关闭连接
 先关闭读，待写发送完毕关闭写
*/
func (this *TCPConn) Close(reason error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	if this.state == 0 || this.state == closed {
		return
	}

	this.closeReason = reason
	this.state = closed
	_ = this.conn.CloseRead()
	// 触发循环
	SendNotifyChan(this.sendNotifyCh)
}

func (this *TCPConn) close() {
	_ = this.conn.Close()
	this.lock.Lock()
	callback := this.opts.CloseCallback
	msg := this.closeReason
	this.lock.Unlock()
	if callback != nil {
		callback(this, msg)
	}
}

type TCPListener struct {
	listener *net.TCPListener
	options  []Option
	started  int32
}

func NewTCPListener(network, addr string, options ...Option) (*TCPListener, error) {
	tcpAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP(tcpAddr.Network(), tcpAddr)
	return &TCPListener{listener: listener, options: options}, err
}

func (l *TCPListener) Listen(newClient func(session Session)) error {
	if newClient == nil {
		return ErrNewClientNil
	}

	if !atomic.CompareAndSwapInt32(&l.started, 0, 1) {
		return ErrStateFailed
	}

	for {
		conn, err := l.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(time.Millisecond * 10)
				continue
			} else {
				return err
			}
		}
		tcpConn, err := NewTCPConn(conn.(*net.TCPConn), l.options...)
		if err != nil {
			return err
		}
		newClient(tcpConn)
	}

}

func (l *TCPListener) Addr() net.Addr {
	return l.listener.Addr()
}

func (l *TCPListener) Close() {
	if atomic.CompareAndSwapInt32(&l.started, 1, 0) {
		_ = l.listener.Close()
	}

}

func DialTCP(network, addr string, timeout time.Duration, options ...Option) (Session, error) {
	tcpAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial(tcpAddr.Network(), addr)
	if err != nil {
		return nil, err
	}

	return NewTCPConn(conn.(*net.TCPConn), options...)
}
