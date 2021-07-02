package dnet

import (
	"fmt"
	"github.com/yddeng/utils/buffer"
	"io"
	"net"
	"reflect"
)

// default编解码器
// 消息 -- 格式: 消息头(消息len), 消息体

const (
	lenSize  = 2       // 消息长度（消息体的长度）
	headSize = lenSize // 消息头长度
	buffSize = 65535   // 缓存容量(与lenSize有关，2字节最大65535）
)

type defTCPCodec struct {
	readBuf  *buffer.Buffer
	dataLen  uint16
	readHead bool
}

func newTCPCodec() *defTCPCodec {
	return &defTCPCodec{
		readBuf: &buffer.Buffer{},
	}
}

//解码
func (decoder *defTCPCodec) Decode(reader io.Reader) (interface{}, error) {
	for {
		msg, err := decoder.unPack()

		if msg != nil {
			return msg, nil

		} else if err == nil {
			_, err1 := decoder.readBuf.ReadFrom(reader)
			if err1 != nil {
				return nil, err1
			}
		} else {
			return nil, err
		}
	}
}

func (decoder *defTCPCodec) unPack() ([]byte, error) {

	if !decoder.readHead {
		if decoder.readBuf.Len() < headSize {
			return nil, nil
		}

		decoder.dataLen, _ = decoder.readBuf.ReadUint16BE()
		decoder.readHead = true
	}

	if decoder.readBuf.Len() < int(decoder.dataLen) {
		return nil, nil
	}

	data, _ := decoder.readBuf.ReadBytes(int(decoder.dataLen))

	decoder.readHead = false
	return data, nil
}

//编码
func (encoder *defTCPCodec) Encode(o interface{}) ([]byte, error) {

	data, ok := o.([]byte)
	if !ok {
		return nil, fmt.Errorf("dnet:Encode interface{} is %s, need type []byte", reflect.TypeOf(o))
	}

	dataLen := len(data)
	if dataLen > buffSize-headSize {
		return nil, fmt.Errorf("dnet:Encode dataLen is too large,len: %d", dataLen)
	}

	msgLen := dataLen + headSize
	buff := buffer.NewBufferWithCap(msgLen)

	//写入data长度
	buff.WriteUint16BE(uint16(dataLen))
	//data数据
	buff.WriteBytes(data)

	return buff.Bytes(), nil
}

// TCPSession
type TCPSession struct {
	*session
}

// NewTCPSession return an initialized *TCPSession
func NewTCPSession(conn net.Conn, options ...Option) *TCPSession {
	op := loadOptions(options...)
	if op.MsgCallback == nil {
		// need message callback
		panic(ErrNilMsgCallBack)
	}
	// init default codec
	if op.Codec == nil {
		op.Codec = newTCPCodec()
	}

	return &TCPSession{
		session: newSession(conn, op),
	}
}

//func (this *TCPSession) CloseRead() error {
//	return this.conn.CloseRead()
//}
//
//func (this *TCPSession) CloseWrite() error {
//	return this.conn.CloseWrite()
//}
