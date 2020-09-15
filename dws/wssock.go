package dws

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/yddeng/dnet"
	"net"
	"reflect"
	"sync"
	"time"
)

const (
	started = 0x01 //0000 0001
	closed  = 0x02 //0000 0010
)

const sendBufChanSize = 1024

type WSConn struct {
	flag         byte
	conn         *websocket.Conn
	ctx          interface{}   //用户数据
	readTimeout  time.Duration // 读超时
	writeTimeout time.Duration // 写超时

	sendBufChan chan []byte //发送队列

	msgCallback   func(interface{}, error)                  //消息回调
	closeCallback func(session dnet.Session, reason string) //关闭连接回调
	closeReason   string                                    //关闭原因

	lock sync.Mutex
}

func NewWSConn(conn *websocket.Conn) *WSConn {
	return &WSConn{
		conn:        conn,
		sendBufChan: make(chan []byte, sendBufChanSize),
	}
}

//读写超时
func (this *WSConn) SetTimeout(readTimeout, writeTimeout time.Duration) {
	defer this.lock.Unlock()
	this.lock.Lock()

	this.readTimeout = readTimeout
	this.writeTimeout = writeTimeout
}

func (this *WSConn) LocalAddr() net.Addr {
	return this.conn.LocalAddr()
}

func (this *WSConn) NetConn() interface{} {
	return this.conn
}

//对端地址
func (this *WSConn) RemoteAddr() net.Addr {
	return this.conn.RemoteAddr()
}

func (this *WSConn) SetCodec(codec dnet.Codec) {}

func (this *WSConn) SetCloseCallBack(closeCallback func(session dnet.Session, reason string)) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.closeCallback = closeCallback
}

func (this *WSConn) SetContext(ctx interface{}) {
	this.lock.Lock()
	this.ctx = ctx
	this.lock.Unlock()
}

func (this *WSConn) Context() interface{} {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.ctx
}

//开启消息处理
func (this *WSConn) Start(msgCb func(interface{}, error)) error {
	if msgCb == nil {
		return dnet.ErrNoMsgCallBack
	}

	this.lock.Lock()
	if this.flag == started {
		return dnet.ErrStateFailed
	}
	this.flag = started
	this.msgCallback = msgCb
	this.lock.Unlock()

	go this.receiveThread()
	go this.sendThread()

	return nil
}

func (this *WSConn) isClose() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.flag == closed
}

//接收线程
func (this *WSConn) receiveThread() {
	for {
		if this.isClose() {
			return
		}
		if this.readTimeout > 0 {
			_ = this.conn.SetReadDeadline(time.Now().Add(this.readTimeout))
		}
		_, msg, err := this.conn.ReadMessage()
		if this.isClose() {
			return
		}
		if err != nil {
			this.msgCallback(nil, err)
		} else {
			this.msgCallback(msg, err)
		}
	}
}

//发送线程
func (this *WSConn) sendThread() {
	defer this.close()
	for {
		data, isOpen := <-this.sendBufChan
		if !isOpen {
			break
		}
		if this.writeTimeout > 0 {
			_ = this.conn.SetWriteDeadline(time.Now().Add(this.writeTimeout))
		}

		err := this.conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			this.msgCallback(nil, err)
		}

	}
}

func (this *WSConn) Send(o interface{}) error {
	if o == nil {
		return dnet.ErrSendMsgNil
	}

	data, ok := o.([]byte)
	if !ok {
		return fmt.Errorf("interface {} is %s,need []byte or use SendMsg(data []byte)", reflect.TypeOf(o).String())
	}

	return this.SendBytes(data)
}

func (this *WSConn) SendBytes(data []byte) error {
	if len(data) == 0 {
		return dnet.ErrSendMsgNil
	}

	//非堵塞
	if len(this.sendBufChan) == sendBufChanSize {
		return dnet.ErrSendChanFull
	}

	this.lock.Lock()
	if this.flag == 0 {
		return dnet.ErrStateFailed
	}
	if this.flag == closed {
		return dnet.ErrStateFailed
	}
	this.lock.Unlock()

	this.sendBufChan <- data
	return nil
}

/*
 主动关闭连接
 先关闭读，待写发送完毕关闭写
*/
func (this *WSConn) Close(reason string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if (this.flag & closed) > 0 {
		return
	}

	close(this.sendBufChan)
	this.closeReason = reason
	this.flag = closed
}

func (this *WSConn) close() {
	_ = this.conn.Close()
	this.lock.Lock()
	callback := this.closeCallback
	msg := this.closeReason
	this.lock.Unlock()
	if callback != nil {
		callback(this, msg)
	}
}
