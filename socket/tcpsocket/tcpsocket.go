package tcpsocket

import (
	"fmt"
	"github.com/yddeng/dnet"
	"net"
	"sync"
	"time"
)

var (
	errSocketStarted = fmt.Errorf("TcpSocket is already started")
	errNotStarted    = fmt.Errorf("TcpSocket is not started")
	errSocketClosed  = fmt.Errorf("TcpSocket is closed")
	errNoCodec       = fmt.Errorf("TcpSocket without codec")
	errNoMsgCallBack = fmt.Errorf("TcpSocket without msgcallback")
	errSendMsgNil    = fmt.Errorf("TcpSocket send msg is nil")
	errSendChanFull  = fmt.Errorf("TcpSocket send chan is full")
)

const (
	started = 0x01 //0000 0001
	rclosed = 0x02 //0000 0010
	wclosed = 0x04 //0000 0100
	closed  = 0x06 //0000 0110
)

const sendChanSize = 1024

type TcpSocket struct {
	flag         byte
	conn         net.Conn
	uData        interface{}   //用户数据
	readTimeout  time.Duration // 读超时
	writeTimeout time.Duration // 写超时

	codec    dnet.Codec  //编解码器
	sendChan chan []byte //发送队列

	msgCallback   func(interface{}, error) //消息回调
	closeCallback func(string)             //关闭连接回调
	closeReason   string                   //关闭原因

	lock sync.Mutex
}

func NewTcpSocket(conn net.Conn) *TcpSocket {
	return &TcpSocket{
		conn:     conn,
		sendChan: make(chan []byte, sendChanSize),
	}
}

//读写超时
func (this *TcpSocket) SetTimeout(readTimeout, writeTimeout time.Duration) {
	defer this.lock.Unlock()
	this.lock.Lock()

	this.readTimeout = readTimeout
	this.writeTimeout = writeTimeout
}

func (this *TcpSocket) LocalAddr() net.Addr {
	return this.conn.LocalAddr()
}

func (this *TcpSocket) NetConn() interface{} {
	return this.conn
}

//对端地址
func (this *TcpSocket) RemoteAddr() net.Addr {
	return this.conn.RemoteAddr()
}

func (this *TcpSocket) SetCodec(codec dnet.Codec) {
	this.lock.Lock()
	this.codec = codec
	this.lock.Unlock()
}

func (this *TcpSocket) SetCloseCallBack(closeCallback func(reason string)) {
	defer this.lock.Unlock()
	this.lock.Lock()
	this.closeCallback = closeCallback
}

func (this *TcpSocket) SetUserData(ud interface{}) {
	this.lock.Lock()
	this.uData = ud
	this.lock.Unlock()
}

func (this *TcpSocket) GetUserData() interface{} {
	defer this.lock.Unlock()
	this.lock.Lock()
	return this.uData
}

//开启消息处理
func (this *TcpSocket) Start(msgCb func(interface{}, error)) error {
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
		return errNoCodec
	}

	this.msgCallback = msgCb

	go this.receiveThread()
	go this.sendThread()

	return nil
}

func (this *TcpSocket) getFlag() byte {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.flag
}

//接收线程
func (this *TcpSocket) receiveThread() {

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
func (this *TcpSocket) sendThread() {

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

func (this *TcpSocket) Send(o interface{}) error {
	if o == nil {
		return errSendMsgNil
	}

	this.lock.Lock()
	if this.codec == nil {
		return errNoCodec
	}
	this.lock.Unlock()

	data, err := this.codec.Encode(o)
	if err != nil {
		return err
	}

	return this.SendBytes(data)
}

func (this *TcpSocket) SendBytes(data []byte) error {
	if len(data) == 0 {
		return errSendMsgNil
	}

	//非堵塞
	if len(this.sendChan) == sendChanSize {
		return errSendChanFull
	}

	this.lock.Lock()
	if (this.flag & started) == 0 {
		return errNotStarted
	}

	if (this.flag & wclosed) > 0 {
		return errSocketClosed
	}
	this.lock.Unlock()

	this.sendChan <- data
	return nil
}

/*
 主动关闭连接
 先关闭读，待写发送完毕关闭写
*/
func (this *TcpSocket) Close(reason string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if (this.flag & closed) > 0 {
		return
	}

	this.closeReason = reason
	this.flag |= rclosed
	close(this.sendChan)
}

func (this *TcpSocket) CloseRead() {
	this.lock.Lock()
	defer this.lock.Unlock()
	if (this.flag & rclosed) > 0 {
		return
	}

	this.flag |= rclosed
	this.conn.(*net.TCPConn).CloseRead()
}

func (this *TcpSocket) close() {
	this.conn.Close()
	this.lock.Lock()
	this.flag |= closed
	this.lock.Unlock()
	if this.closeCallback != nil {
		this.closeCallback(this.closeReason)
	}
}
