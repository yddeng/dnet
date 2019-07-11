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
	uData        interface{}
	readTimeout  time.Duration
	writeTimeout time.Duration

	reader   dnet.Reader
	encode   dnet.Encode
	sendChan chan interface{}

	callback func(interface{})

	closeChan chan bool //当前连接关闭
	lock      sync.Mutex
}

func NewSession(conn net.Conn, e dnet.Encode, r dnet.Reader) *Session {
	return &Session{
		conn:      conn,
		uData:     nil,
		encode:    e,
		reader:    r,
		sendChan:  make(chan interface{}, sendChanSize),
		closeChan: make(chan bool),
	}
}

func (this *Session) SetTimeout(readTimeout, writeTimeout time.Duration) {
	defer this.lock.Unlock()
	this.lock.Lock()

	this.readTimeout = readTimeout
	this.writeTimeout = writeTimeout
}

func (this *Session) GetRemoteAddr() string {
	return this.conn.RemoteAddr().String()
}

func (this *Session) SetUserData(ud interface{}) {
	this.uData = ud
}

func (this *Session) GetUserData() interface{} {
	return this.uData
}

func (this *Session) SetEncode(e dnet.Encode) {
	this.encode = e
}

func (this *Session) SetReader(r dnet.Reader) {
	this.reader = r
}

func (this *Session) Start(cb func(interface{})) {

	if this.encode == nil || this.reader == nil {
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

		msg, err := this.reader.Receive(this.conn)

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

			data, err := this.encode.Pack(msg)
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

func (this *Session) Send(data interface{}) {
	this.sendChan <- data
}

func (this *Session) Close() error {
	defer close(this.closeChan)
	return this.conn.Close()
}
