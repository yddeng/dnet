package socket

import (
	"errors"
	"github.com/tagDong/dnet/util"
	"go-common/library/log"
	"io"
	"net"
	"sync"
	"time"
)

var (
	ErrServerClose  = errors.New("server Close")
	ErrSessionClose = errors.New("session Close")
)

const (
	sendChSize = 1024
)

type Session struct {
	conn         net.Conn
	uData        interface{}
	readTimeout  time.Duration
	writeTimeout time.Duration

	receiver ReceiverI
	sendChan chan []byte

	callback func(interface{})

	servCloseCh chan bool //服务器器关闭
	closeChan   chan bool //当前连接关闭
	lock        sync.Mutex
}

func NewSession(conn net.Conn, serverCloseChan chan bool) *Session {
	return &Session{
		conn:        conn,
		uData:       nil,
		receiver:    NewReceiver(),
		sendChan:    make(chan []byte, sendChSize),
		servCloseCh: serverCloseChan,
		closeChan:   make(chan bool),
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

func (this *Session) Start(cb func(interface{})) {

	if err := this.isClose(); err != nil {
		log.Info(err.Error())
		return
	}

	this.callback = cb

	go this.recvMsg()
	go this.sendMsg()
}

func (this *Session) isClose() error {
	select {
	case <-this.servCloseCh:
		return ErrServerClose
	case <-this.closeChan:
		return ErrSessionClose
	default:
		return nil
	}
}

func (this *Session) recvMsg() {
	for {
		if err := this.isClose(); err != nil {
			return
		}

		if this.readTimeout > 0 {
			this.conn.SetReadDeadline(time.Now().Add(this.readTimeout))
		}

		msg, err := this.receiver.ReadAndUnPack(this.conn)

		if err != nil {
			if err == io.EOF {
				log.Info("read Close")
			} else {
				log.Info("read err: ", err.Error())
			}
			this.Close()
			return
		}

		if err := this.isClose(); err != nil {
			return
		}

		if msg != nil {
			this.callback(msg)
		}
	}
}

func (this *Session) sendMsg() {
	for {
		select {
		case <-this.servCloseCh:
			return
		case <-this.closeChan:
			return
		case msg := <-this.sendChan:
			if this.writeTimeout > 0 {
				this.conn.SetWriteDeadline(time.Now().Add(this.writeTimeout))
			}

			_, err := this.conn.Write(msg)
			if err != nil {
				this.Close()
			}
		}

	}
}

func (this *Session) Send(msg []byte) {
	buff := util.NewBuffer(lenSize + len(msg))
	buff.PutUint32(uint32(len(msg)))
	buff.Write(msg)

	this.sendChan <- buff.Buff()
}

func (this *Session) Close() error {
	defer close(this.closeChan)
	return this.conn.Close()
}
