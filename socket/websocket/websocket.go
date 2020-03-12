package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/yddeng/dnet"
	"net"
	"reflect"
	"sync"
	"time"
)

var (
	errSocketStarted = fmt.Errorf("WebSocket is already started")
	errNotStarted    = fmt.Errorf("WebSocket is not started")
	errSocketClosed  = fmt.Errorf("WebSocket is closed")
	errNoCodec       = fmt.Errorf("WebSocket without codec")
	errNoMsgCallBack = fmt.Errorf("WebSocket without msgcallback")
	errSendMsgNil    = fmt.Errorf("WebSocket send msg is nil")
	errSendChanFull  = fmt.Errorf("WebSocket send chan is full")
)

const (
	started = 0x01 //0000 0001
	rclosed = 0x02 //0000 0010
	wclosed = 0x04 //0000 0100
	closed  = 0x06 //0000 0110
)

const sendChanSize = 1024

type WebSocket struct {
	flag         byte
	conn         *websocket.Conn
	uData        interface{}   //用户数据
	readTimeout  time.Duration // 读超时
	writeTimeout time.Duration // 写超时

	sendChan      chan []byte              //发送队列
	msgCallback   func(interface{}, error) //消息回调
	closeCallback func(string)             //关闭连接回调
	closeReason   string                   //关闭原因

	lock sync.Mutex
}

func NewWebSocket(conn *websocket.Conn) *WebSocket {
	return &WebSocket{
		conn:     conn,
		sendChan: make(chan []byte, sendChanSize),
	}
}

//读写超时
func (this *WebSocket) SetTimeout(readTimeout, writeTimeout time.Duration) {
	defer this.lock.Unlock()
	this.lock.Lock()

	this.readTimeout = readTimeout
	this.writeTimeout = writeTimeout
}

func (this *WebSocket) LocalAddr() net.Addr {
	return this.conn.LocalAddr()
}

func (this *WebSocket) NetConn() interface{} {
	return this.conn
}

//对端地址
func (this *WebSocket) RemoteAddr() net.Addr {
	return this.conn.RemoteAddr()
}

func (this *WebSocket) SetCodec(codec dnet.Codec) {}

func (this *WebSocket) SetCloseCallBack(closeCallback func(reason string)) {
	defer this.lock.Unlock()
	this.lock.Lock()
	this.closeCallback = closeCallback
}

func (this *WebSocket) SetUserData(ud interface{}) {
	this.lock.Lock()
	this.uData = ud
	this.lock.Unlock()
}

func (this *WebSocket) GetUserData() interface{} {
	defer this.lock.Unlock()
	this.lock.Lock()
	return this.uData
}

//开启消息处理
func (this *WebSocket) Start(msgCb func(interface{}, error)) error {
	if msgCb == nil {
		return errNoMsgCallBack
	}

	this.lock.Lock()
	defer this.lock.Unlock()
	if (this.flag & started) > 0 {
		return errSocketStarted
	}
	this.flag |= started

	this.msgCallback = msgCb

	go this.receiveThread()
	go this.sendThread()

	return nil
}

func (this *WebSocket) getFlag() byte {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.flag
}

//接收线程
func (this *WebSocket) receiveThread() {

	for (this.getFlag() & rclosed) == 0 {
		if this.readTimeout > 0 {
			_ = this.conn.SetReadDeadline(time.Now().Add(this.readTimeout))
		}
		_, msg, err := this.conn.ReadMessage()
		if err != nil {
			this.msgCallback(nil, err)
		} else {
			this.msgCallback(msg, err)
		}
	}
}

//发送线程
func (this *WebSocket) sendThread() {

	defer this.close()
	for (this.getFlag() & wclosed) == 0 {

		data, isOpen := <-this.sendChan
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

func (this *WebSocket) Send(o interface{}) error {
	if o == nil {
		return errSendMsgNil
	}

	data, ok := o.([]byte)
	if !ok {
		return fmt.Errorf("interface {} is %s,need []byte or use SendMsg(data []byte)", reflect.TypeOf(o).String())
	}

	return this.SendBytes(data)
}

func (this *WebSocket) SendBytes(data []byte) error {
	if len(data) == 0 {
		return errSendMsgNil
	}

	//非堵塞
	if len(this.sendChan) == sendChanSize {
		return errSendChanFull
	}

	this.lock.Lock()
	defer this.lock.Unlock()
	if (this.flag & started) == 0 {
		return errNotStarted
	}

	if (this.flag & wclosed) > 0 {
		return errSocketClosed
	}

	this.sendChan <- data
	return nil
}

/*
 主动关闭连接
 先关闭读，待写发送完毕关闭写
*/
func (this *WebSocket) Close(reason string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if (this.flag & closed) > 0 {
		return
	}

	this.closeReason = reason
	this.flag |= rclosed
	close(this.sendChan)
}

func (this *WebSocket) close() {
	this.conn.Close()
	this.lock.Lock()
	this.flag |= closed
	this.lock.Unlock()
	if this.closeCallback != nil {
		this.closeCallback(this.closeReason)
	}
}
