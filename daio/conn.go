package daio

import (
	"bytes"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/daio/poller"
	"io"
	"net"
	"syscall"

	"sync"
	"time"
)

const (
	started   = 0x01      //0000 0001
	closed    = 0x02      //0000 0010
	streamLen = 64 * 1024 // 每次发送的最大流数据
)

type AioConn struct {
	fd   int
	conn net.Conn
	loop *EventLoop

	flag byte
	ctx  interface{} //用户数据 context

	lastReadTime time.Time
	readTimeout  time.Duration // 读超时
	writeTimeout time.Duration // 写超时

	codec       dnet.Codec   //编解码器
	writeBuffer bytes.Buffer //发送buffer

	msgCallback   func(interface{}, error)                  //消息回调
	closeCallback func(session dnet.Session, reason string) //关闭连接回调

	lock sync.Mutex
}

func newAioConn(fd int, conn net.Conn, eventLoop *EventLoop) *AioConn {
	c := &AioConn{
		fd:          fd,
		conn:        conn,
		loop:        eventLoop,
		writeBuffer: bytes.Buffer{},
	}

	return c
}

func (c *AioConn) handleEvent(fd int, events poller.Event) error {

	if events&poller.EventErr != 0 {
		c.handleClose(io.EOF.Error())
		return nil
	}

	if events&poller.EventWrite != 0 {
		c.handleWrite(fd)
	}
	if events&poller.EventRead != 0 {
		c.handleRead()
	}

	return nil
}

func (c *AioConn) handleClose(reason string) {

	c.lock.Lock()
	if (c.flag & closed) > 0 {
		c.lock.Unlock()
		return
	}
	c.flag = closed
	c.lock.Unlock()

	c.loop.Remove(c.fd)
	_ = syscall.Close(c.fd)

	if c.closeCallback != nil {
		c.closeCallback(c, reason)
	}

}

// io.Reader
func (c *AioConn) Read(buf []byte) (n int, err error) {
	n, err = syscall.Read(c.fd, buf)
	if err != nil {
		n = 0
		if err == syscall.EAGAIN { // 表示没有数据可读
			err = nil
			return
		}
		return
	}
	return
}

func (c *AioConn) handleRead() {
	for {
		msg, err := c.codec.Decode(c)
		if err != nil {
			//关闭连接
			c.msgCallback(nil, err)
		} else {
			if msg != nil {
				c.msgCallback(msg, nil)
			} else {
				return
			}
		}
	}
}

func (c *AioConn) handleWrite(fd int) {
	c.lock.Lock()
	if c.flag == closed {
		c.lock.Unlock()
		return
	}

	data := c.writeBuffer.Next(streamLen)

	n, err := syscall.Write(c.fd, data)
	if err != nil {
		if err == syscall.EAGAIN {
			c.lock.Unlock()
			return
		}
		c.lock.Unlock()
		c.handleClose(err.Error())
		return
	}

	c.writeBuffer.Reset()
	if n < len(data) { // 没有发送完
		c.writeBuffer.Write(data[n:])
	}

	if c.writeBuffer.Len() == 0 {
		_ = c.loop.poll.ModRead(c.fd)
	}

	c.lock.Unlock()

	//fmt.Println("send", n, data)

}

//读写超时
func (this *AioConn) SetTimeout(readTimeout, writeTimeout time.Duration) {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.readTimeout = readTimeout
	this.writeTimeout = writeTimeout
}

func (this *AioConn) NetConn() interface{} { return this.conn }
func (this *AioConn) LocalAddr() net.Addr  { return this.conn.LocalAddr() }
func (this *AioConn) RemoteAddr() net.Addr { return this.conn.RemoteAddr() }

func (this *AioConn) SetCodec(codec dnet.Codec) {
	this.lock.Lock()
	this.codec = codec
	this.lock.Unlock()
}

func (this *AioConn) SetCloseCallBack(closeCallback func(session dnet.Session, reason string)) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.closeCallback = closeCallback
}

func (this *AioConn) SetContext(ctx interface{}) {
	this.lock.Lock()
	this.ctx = ctx
	this.lock.Unlock()
}

func (this *AioConn) Context() interface{} {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.ctx
}

//开启消息处理
func (this *AioConn) Start(msgCb func(interface{}, error)) error {
	if msgCb == nil {
		return dnet.ErrNoMsgCallBack
	}

	this.lock.Lock()
	if this.flag == started {
		this.lock.Unlock()
		return dnet.ErrStateFailed
	}
	this.flag = started

	if this.codec == nil {
		this.lock.Unlock()
		return dnet.ErrNoCodec
	}

	this.msgCallback = msgCb
	this.lock.Unlock()

	return this.loop.Watch(this)
}

func (this *AioConn) Send(o interface{}) error {
	if o == nil {
		return dnet.ErrSendMsgNil
	}

	this.lock.Lock()
	if this.codec == nil {
		this.lock.Unlock()
		return dnet.ErrNoCodec
	}
	codec := this.codec
	this.lock.Unlock()

	data, err := codec.Encode(o)
	if err != nil {
		return err
	}

	return this.SendBytes(data)
}

func (this *AioConn) SendBytes(data []byte) error {
	if len(data) == 0 {
		return dnet.ErrSendMsgNil
	}

	this.lock.Lock()

	if this.flag == 0 {
		this.lock.Unlock()
		return dnet.ErrStateFailed
	}
	if this.flag == closed {
		this.lock.Unlock()
		return dnet.ErrStateFailed
	}
	this.lock.Unlock()

	this.loop.Do(func() {
		_, _ = this.writeBuffer.Write(data)
		_ = this.loop.poll.ModReadWrite(this.fd)
	})

	return nil

}

/*
 主动关闭连接
 先关闭读，待写发送完毕关闭写
*/
func (this *AioConn) Close(reason string) {
	this.lock.Lock()
	if this.flag == closed {
		this.lock.Unlock()
		return
	}
	this.lock.Unlock()

	this.loop.Do(func() {
		this.handleClose(reason)
	})

}
