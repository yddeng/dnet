package dkcp

import (
	"fmt"
	"github.com/xtaci/kcp-go"
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

type KCPConn struct {
	state        byte
	conn         *kcp.UDPSession
	ctx          interface{}   // 用户数据
	readTimeout  time.Duration // 读超时
	writeTimeout time.Duration // 写超时

	codec dnet.Codec // 编解码器

	sendNotifyCh  chan struct{} // 发送消息通知
	sendMessageCh chan *message // 发送队列

	msgCallback func(interface{}, error) // 消息回调

	closeCallback func(session dnet.Session, reason string) // 关闭连接回调
	closeReason   string                                    // 关闭原因

	lock sync.Mutex
}

type message struct {
	needEncode bool
	data       interface{}
}

func NewKCPConn(conn *kcp.UDPSession) *KCPConn {
	return &KCPConn{
		conn:          conn,
		sendNotifyCh:  make(chan struct{}, 1),
		sendMessageCh: make(chan *message, sendBufChanSize),
	}
}

//读写超时
func (this *KCPConn) SetTimeout(readTimeout, writeTimeout time.Duration) {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.readTimeout = readTimeout
	this.writeTimeout = writeTimeout
}

func (this *KCPConn) LocalAddr() net.Addr {
	return this.conn.LocalAddr()
}

func (this *KCPConn) NetConn() interface{} {
	return this.conn
}

//对端地址
func (this *KCPConn) RemoteAddr() net.Addr {
	return this.conn.RemoteAddr()
}

func (this *KCPConn) SetCodec(codec dnet.Codec) {
	this.lock.Lock()
	this.codec = codec
	this.lock.Unlock()
}

func (this *KCPConn) SetCloseCallBack(closeCallback func(session dnet.Session, reason string)) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.closeCallback = closeCallback
}

func (this *KCPConn) SetContext(ctx interface{}) {
	this.lock.Lock()
	this.ctx = ctx
	this.lock.Unlock()
}

func (this *KCPConn) Context() interface{} {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.ctx
}

func (this *KCPConn) getCodec() dnet.Codec {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.codec
}

func (this *KCPConn) getTimeout() (readTimeout, writeTimeout time.Duration) {
	this.lock.Lock()
	readTimeout, writeTimeout = this.readTimeout, this.writeTimeout
	this.lock.Unlock()
	return
}

//开启消息处理
func (this *KCPConn) Start(msgCb func(interface{}, error)) error {
	if msgCb == nil {
		return dnet.ErrNoMsgCallBack
	}

	this.lock.Lock()
	if this.state == started {
		this.lock.Unlock()
		return dnet.ErrStateFailed
	}
	this.state = started

	if this.codec == nil {
		this.lock.Unlock()
		return dnet.ErrNoCodec
	}

	this.msgCallback = msgCb
	this.lock.Unlock()

	go this.receiveThread()
	go this.sendThread()

	return nil
}

func (this *KCPConn) isClose() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.state == closed
}

//接收线程
func (this *KCPConn) receiveThread() {
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
func (this *KCPConn) sendThread() {
	for {
		select {
		case msg := <-this.sendMessageCh:
			var err error
			var data []byte

			if msg.needEncode {
				// 需要编码的消息
				codec := this.getCodec()
				if codec == nil {
					this.msgCallback(nil, dnet.ErrNoCodec)
					break
				}

				data, err = codec.Encode(msg.data)
				if err != nil {
					this.msgCallback(nil, err)
					break
				}
			} else {
				var ok bool
				data, ok = msg.data.([]byte)
				if !ok {
					this.msgCallback(nil, fmt.Errorf("dkcp: sendTread reflect failed, data is not []byte"))
					break
				}
			}

			if data != nil && len(data) != 0 {
				// 发送的消息
				_, writeTimeout := this.getTimeout()
				if writeTimeout > 0 {
					this.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
				}

				_, err = this.conn.Write(data)
				if err != nil {
					this.msgCallback(nil, err)
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

func (this *KCPConn) Send(o interface{}) error {
	if o == nil {
		return dnet.ErrSendMsgNil
	}

	this.lock.Lock()
	if this.state != started {
		this.lock.Unlock()
		return dnet.ErrStateFailed
	}
	this.lock.Unlock()

	if this.getCodec() == nil {
		return dnet.ErrNoCodec
	}

	//非堵塞
	//if len(this.sendMessageCh) == sendBufChanSize {
	//	return dnet.ErrSendChanFull
	//}

	this.sendMessageCh <- &message{
		needEncode: true,
		data:       o,
	}
	dnet.SendNotifyChan(this.sendNotifyCh)
	return nil
}

func (this *KCPConn) SendBytes(data []byte) error {
	if len(data) == 0 {
		return dnet.ErrSendMsgNil
	}

	this.lock.Lock()
	if this.state != started {
		this.lock.Unlock()
		return dnet.ErrStateFailed
	}
	this.lock.Unlock()

	//非堵塞
	//if len(this.sendMessageCh) == sendBufChanSize {
	//	return dnet.ErrSendChanFull
	//}

	this.sendMessageCh <- &message{
		needEncode: false,
		data:       data,
	}
	dnet.SendNotifyChan(this.sendNotifyCh)
	return nil
}

/*
 主动关闭连接
 先关闭读，待写发送完毕关闭写
*/
func (this *KCPConn) Close(reason string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	if this.state == 0 || this.state == closed {
		return
	}

	this.closeReason = reason
	this.state = closed
	this.conn.Close()
	// 触发循环
	dnet.SendNotifyChan(this.sendNotifyCh)
}

func (this *KCPConn) close() {
	_ = this.conn.Close()
	this.lock.Lock()
	callback := this.closeCallback
	msg := this.closeReason
	this.lock.Unlock()
	if callback != nil {
		callback(this, msg)
	}
}
