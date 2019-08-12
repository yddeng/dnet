package dnet

import (
	"fmt"
	"net"
	"sync"
	"time"
)

var (
	errSocketStarted = fmt.Errorf("Socket is already started")
	errNotStarted    = fmt.Errorf("Socket is not started")
	errSocketClosed  = fmt.Errorf("Socket is closed")
	errNoCodec       = fmt.Errorf("Socket without codec")
	errNoMsgCallBack = fmt.Errorf("Socket without msgcallback")
	errSendMsgNil    = fmt.Errorf("Socket send msg is nil")
	errSendChanFull  = fmt.Errorf("Socket send chan is full")
)

const (
	started = 0x01 //0000 0001
	rclosed = 0x02 //0000 0010
	wclosed = 0x04 //0000 0100
	closed  = 0x06 //0000 0110
)

const sendChanSize = 1024

type Socket struct {
	flag         byte
	conn         net.Conn
	uData        interface{}   //用户数据
	readTimeout  time.Duration // 读超时
	writeTimeout time.Duration // 写超时

	codec    Codec       //编解码器
	sendChan chan []byte //发送队列

	msgCallback   func(interface{}, error) //消息回调
	closeCallback func(string)             //关闭连接回调
	closeReason   string                   //关闭原因

	lock sync.Mutex
}

func NewSocket(conn net.Conn) *Socket {
	return &Socket{
		conn:     conn,
		sendChan: make(chan []byte, sendChanSize),
	}
}

//读写超时
func (this *Socket) SetTimeout(readTimeout, writeTimeout time.Duration) {
	defer this.lock.Unlock()
	this.lock.Lock()

	this.readTimeout = readTimeout
	this.writeTimeout = writeTimeout
}

func (this *Socket) LocalAddr() net.Addr {
	return this.conn.LocalAddr()
}

//对端地址
func (this *Socket) RemoteAddr() net.Addr {
	return this.conn.RemoteAddr()
}

func (this *Socket) SetCodec(codec Codec) {
	this.lock.Lock()
	this.codec = codec
	this.lock.Unlock()
}

func (this *Socket) SetCloseCallBack(closeCallback func(reason string)) {
	defer this.lock.Unlock()
	this.lock.Lock()
	this.closeCallback = closeCallback
}

func (this *Socket) SetUserData(ud interface{}) {
	this.lock.Lock()
	this.uData = ud
	this.lock.Unlock()
}

func (this *Socket) GetUserData() interface{} {
	defer this.lock.Unlock()
	this.lock.Lock()
	return this.uData
}

//开启消息处理
func (this *Socket) Start(msgCb func(interface{}, error)) error {
	if msgCb == nil {
		return errNoMsgCallBack
	}

	this.lock.Lock()
	defer this.lock.Unlock()
	if (this.flag & started) > 0 {
		return errSocketStarted
	}
	this.flag |= started

	if this.codec == nil {
		this.codec = NewDefCodec()
	}
	this.msgCallback = msgCb

	go this.receiveRoutine()
	go this.sendRoutine()

	return nil
}

func (this *Socket) getFlag() byte {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.flag
}

//接收线程
func (this *Socket) receiveRoutine() {

	for (this.getFlag() & rclosed) == 0 {

		if this.readTimeout > 0 {
			this.conn.SetReadDeadline(time.Now().Add(this.readTimeout))
		}

		msg, err := this.codec.Decode(this.conn)

		if err != nil {
			//if err == io.EOF {
			//	log.Println("read Close", err.Error())
			//} else {
			//	log.Println("read err: ", err.Error())
			//}
			//关闭连接
			//this.Close(err.Error())
			this.msgCallback(nil, err)
		} else {
			if msg != nil {
				this.msgCallback(msg, nil)
			}
		}
	}
}

//发送线程
func (this *Socket) sendRoutine() {

	defer this.close()
	for (this.getFlag() & wclosed) == 0 {

		data, isOpen := <-this.sendChan
		if !isOpen {
			break
		}
		if this.writeTimeout > 0 {
			this.conn.SetWriteDeadline(time.Now().Add(this.writeTimeout))
		}

		_, err := this.conn.Write(data)
		if err != nil {
			//log.Println("write err: ", err.Error())
			//this.Close(err.Error())
			this.msgCallback(nil, err)
		}

	}
}

func (this *Socket) Send(o interface{}) error {
	if o == nil {
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

	if this.codec == nil {
		return errNoCodec
	}

	data, err := this.codec.Encode(o)
	if err != nil {
		return err
	}

	this.sendChan <- data

	return nil
}

/*
 主动关闭连接
 先关闭读，待写发送完毕关闭写
*/
func (this *Socket) Close(reason string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if (this.flag & closed) > 0 {
		return
	}

	this.closeReason = reason
	this.flag |= rclosed
	close(this.sendChan)
}

func (this *Socket) close() {
	this.conn.Close()
	this.lock.Lock()
	defer this.lock.Unlock()
	this.flag |= closed
	if this.closeCallback != nil {
		this.closeCallback(this.closeReason)
	}
}
