package socket

import (
	"errors"
	"github.com/tagDong/dnet"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

var (
	ErrServerClose  = errors.New("server Close")
	ErrSessionClose = errors.New("session Close")
)

const sendChanSize = 1024

type Session struct {
	conn         net.Conn
	uData        interface{}   //用户数据
	readTimeout  time.Duration // 读取超时
	writeTimeout time.Duration // 发送超时

	codec    dnet.Codec        //编解码器
	sendChan chan dnet.Message //发送队列

	callback func(interface{}) //消息回调

	closeChan chan bool //当前连接关闭
	lock      sync.Mutex
}

func NewSession(conn net.Conn, codec dnet.Codec) *Session {
	return &Session{
		conn:      conn,
		uData:     nil,
		codec:     codec,
		sendChan:  make(chan dnet.Message, sendChanSize),
		closeChan: make(chan bool),
	}
}

//读写超时
func (this *Session) SetTimeout(readTimeout, writeTimeout time.Duration) {
	defer this.lock.Unlock()
	this.lock.Lock()

	this.readTimeout = readTimeout
	this.writeTimeout = writeTimeout
}

//对端地址
func (this *Session) GetRemoteAddr() string {
	return this.conn.RemoteAddr().String()
}

func (this *Session) SetUserData(ud interface{}) {
	this.uData = ud
}

func (this *Session) GetUserData() interface{} {
	return this.uData
}

//开启消息处理
func (this *Session) Start(cb func(interface{})) {

	if this.codec == nil {
		return
	}

	if cb == nil {
		return
	}

	this.callback = cb

	go this.recvMsg()
	go this.sendMsg()
}

//接收线程
func (this *Session) recvMsg() {
	defer this.Close()
	for {
		select {
		case <-this.closeChan:
			return
		default:
		}

		if this.readTimeout > 0 {
			this.conn.SetReadDeadline(time.Now().Add(this.readTimeout))
		}

		msg, err := this.codec.Decode(this.conn)

		if err != nil {
			if err == io.EOF {
				log.Println("read Close")
			} else {
				log.Println("read err: ", err.Error())
			}
			return
		}

		if msg != nil {
			this.callback(msg)
		}
	}
}

//发送线程
func (this *Session) sendMsg() {
	for {
		select {
		case <-this.closeChan:
			return
		case msg := <-this.sendChan:

			if this.writeTimeout > 0 {
				this.conn.SetWriteDeadline(time.Now().Add(this.writeTimeout))
			}

			data, err := this.codec.Encode(msg.(dnet.Message))
			if err != nil {
				this.Close()
			}

			_, err = this.conn.Write(data)
			if err != nil {
				this.Close()
			}
		}

	}
}

func (this *Session) Send(msg dnet.Message) {
	this.sendChan <- msg
}

func (this *Session) Close() error {
	defer close(this.closeChan)
	return this.conn.Close()
}
