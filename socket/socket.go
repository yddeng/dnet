package socket

import (
	"fmt"
	"github.com/tagDong/dnet"
	"github.com/tagDong/dnet/util"
	"io"
	"net"
	"sync"
	"time"
)

const sizeLen = 2

type StreamSocket struct {
	conn         net.Conn
	uData        interface{}
	readTimeout  time.Duration
	writeTimeout time.Duration
	readBuffer   *util.Buffer
	writeBuffer  *util.Buffer
	lock         sync.Mutex
}

func NewStreamSocket(conn net.Conn) dnet.StreamSession {
	return &StreamSocket{
		conn:        conn,
		uData:       nil,
		readBuffer:  util.NewBuffer(),
		writeBuffer: util.NewBuffer(),
	}
}

func (this *StreamSocket) SetTimeout(readTimeout, writeTimeout time.Duration) {
	defer this.lock.Unlock()
	this.lock.Lock()

	this.readTimeout = readTimeout
	this.writeTimeout = writeTimeout
}

func (this *StreamSocket) GetRemoteAddr() string {
	return this.conn.RemoteAddr().String()
}

func (this *StreamSocket) SetUserData(ud interface{}) {
	this.uData = ud
}

func (this *StreamSocket) GetUserData() interface{} {
	return this.uData
}

func (this *StreamSocket) StartReceive(cb func([]byte)) {
	if cb == nil {
		return
	}
	go this.readGoroutine(cb)
	//go this.writeGoroutine()
}

func (this *StreamSocket) readGoroutine(cb func([]byte)) {
	for {
		if this.readTimeout > 0 {
			this.conn.SetReadDeadline(time.Now().Add(this.readTimeout))
		}

		count, err := this.readBuffer.Reader(this.conn)
		if count > 0 {
			unPack(this.readBuffer, cb)
		}
		if err != nil {
			if err == io.EOF {
				fmt.Println("read Close")
			} else {
				fmt.Println("read err", err)
			}
			this.Close()
			return
		}
	}
}

func unPack(buf *util.Buffer, cb func([]byte)) {
	for buf.Size() != 0 {
		msglen, err := buf.GetUint16()
		if err != nil || buf.Size() < int(msglen)+sizeLen {
			return
		}

		cb(buf.GetBytes(sizeLen, int(msglen)))
		buf.Reset(sizeLen + int(msglen))
	}
}

func (this *StreamSocket) writeGoroutine() {

}

func (this *StreamSocket) Send(msg []byte) {
	this.writeBuffer.AppendUint16(uint16(len(msg)))
	this.writeBuffer.AppendByte(msg)

	for this.writeBuffer.Size() != 0 {
		if this.writeTimeout > 0 {
			this.conn.SetWriteDeadline(time.Now().Add(this.writeTimeout))
		}

		n, err := this.conn.Write(this.writeBuffer.GetBytes(0, this.writeBuffer.Size()))
		if err != nil {
			fmt.Println("send err", err)
		}
		this.writeBuffer.Reset(n)
	}
}

func (this *StreamSocket) Close() error {
	return this.conn.Close()
}
