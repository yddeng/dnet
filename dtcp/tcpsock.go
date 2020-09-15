package dtcp

import (
	"github.com/yddeng/dnet"
	"net"
	"sync"
	"time"
)

const (
	started = 0x01 //0000 0001
	closed  = 0x02 //0000 0010
)

const sendBufChanSize = 1024

type TCPConn struct {
	state        byte
	conn         *net.TCPConn
	ctx          interface{}   // 用户数据
	readTimeout  time.Duration // 读超时
	writeTimeout time.Duration // 写超时

	codec dnet.Codec // 编解码器

	sendNotifyCh  chan struct{}    // 发送消息通知
	sendBufferCh  chan []byte      // 发送队列
	pendingEncode chan interface{} // 待编码发送的消息

	msgCallback func(interface{}, error) // 消息回调

	closeCallback func(session dnet.Session, reason string) // 关闭连接回调
	closeReason   string                                    // 关闭原因

	lock sync.Mutex
}

func NewTCPConn(conn *net.TCPConn) *TCPConn {
	return &TCPConn{
		conn:          conn,
		sendNotifyCh:  make(chan struct{}, 1),
		sendBufferCh:  make(chan []byte, sendBufChanSize),
		pendingEncode: make(chan interface{}, sendBufChanSize),
	}
}

//读写超时
func (this *TCPConn) SetTimeout(readTimeout, writeTimeout time.Duration) {
	defer this.lock.Unlock()
	this.lock.Lock()

	this.readTimeout = readTimeout
	this.writeTimeout = writeTimeout
}

func (this *TCPConn) LocalAddr() net.Addr {
	return this.conn.LocalAddr()
}

func (this *TCPConn) NetConn() interface{} {
	return this.conn
}

//对端地址
func (this *TCPConn) RemoteAddr() net.Addr {
	return this.conn.RemoteAddr()
}

func (this *TCPConn) SetCodec(codec dnet.Codec) {
	this.lock.Lock()
	this.codec = codec
	this.lock.Unlock()
}

func (this *TCPConn) SetCloseCallBack(closeCallback func(session dnet.Session, reason string)) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.closeCallback = closeCallback
}

func (this *TCPConn) SetContext(ctx interface{}) {
	this.lock.Lock()
	this.ctx = ctx
	this.lock.Unlock()
}

func (this *TCPConn) Context() interface{} {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.ctx
}

func (this *TCPConn) getCodec() dnet.Codec {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.codec
}

func (this *TCPConn) getTimeout() (readTimeout, writeTimeout time.Duration) {
	this.lock.Lock()
	readTimeout, writeTimeout = this.readTimeout, this.writeTimeout
	this.lock.Unlock()
	return
}

//开启消息处理
func (this *TCPConn) Start(msgCb func(interface{}, error)) error {
	if msgCb == nil {
		return dnet.ErrNoMsgCallBack
	}

	this.lock.Lock()
	if this.state == started {
		return dnet.ErrStateFailed
	}
	this.state = started

	if this.codec == nil {
		return dnet.ErrNoCodec
	}

	this.msgCallback = msgCb
	this.lock.Unlock()

	go this.receiveThread()
	go this.sendThread()

	return nil
}

func (this *TCPConn) isClose() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.state == closed
}

//接收线程
func (this *TCPConn) receiveThread() {
	for {
		if this.isClose() {
			return
		}

		readTimeout, _ := this.getTimeout()
		if readTimeout > 0 {
			this.conn.SetReadDeadline(time.Now().Add(readTimeout))
		}

		codec := this.getCodec()
		if codec != nil {
			msg, err := codec.Decode(this.conn)
			if this.isClose() {
				return
			}
			if err != nil {
				this.msgCallback(nil, err)
			} else {
				if msg != nil {
					this.msgCallback(msg, nil)
				}
			}
		} else {
			this.msgCallback(nil, dnet.ErrNoCodec)
		}
	}
}

// 发送线程
// 关闭连接时，发送完后再关闭
func (this *TCPConn) sendThread() {
	for {
		select {
		case o := <-this.pendingEncode:
			// 需要编码的消息
			codec := this.getCodec()
			if codec == nil {
				this.msgCallback(nil, dnet.ErrNoCodec)
				break
			}

			data, err := codec.Encode(o)
			if err != nil {
				this.msgCallback(nil, err)
				break
			}

			this.sendBufferCh <- data
			dnet.SendNotifyChan(this.sendNotifyCh)

		case data := <-this.sendBufferCh:
			// 直接发送的消息
			_, writeTimeout := this.getTimeout()
			if writeTimeout > 0 {
				this.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			}

			_, err := this.conn.Write(data)
			if err != nil {
				this.msgCallback(nil, err)
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
		return dnet.ErrSendMsgNil
	}

	this.lock.Lock()
	if this.state != started {
		return dnet.ErrStateFailed
	}
	this.lock.Unlock()

	if this.getCodec() == nil {
		return dnet.ErrNoCodec
	}

	this.pendingEncode <- o
	dnet.SendNotifyChan(this.sendNotifyCh)
	return nil
}

func (this *TCPConn) SendBytes(data []byte) error {
	if len(data) == 0 {
		return dnet.ErrSendMsgNil
	}

	this.lock.Lock()
	if this.state != started {
		return dnet.ErrStateFailed
	}
	this.lock.Unlock()

	//非堵塞
	if len(this.sendBufferCh) == sendBufChanSize {
		return dnet.ErrSendChanFull
	}

	this.sendBufferCh <- data
	dnet.SendNotifyChan(this.sendNotifyCh)
	return nil
}

/*
 主动关闭连接
 先关闭读，待写发送完毕关闭写
*/
func (this *TCPConn) Close(reason string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.closeReason = reason
	if this.state == 0 || this.state == closed {
		return
	}

	this.state = closed
	this.conn.CloseRead()
	// 触发循环
	dnet.SendNotifyChan(this.sendNotifyCh)
}

func (this *TCPConn) close() {
	_ = this.conn.Close()
	this.lock.Lock()
	callback := this.closeCallback
	msg := this.closeReason
	this.lock.Unlock()
	if callback != nil {
		callback(this, msg)
	}
}
