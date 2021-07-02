package dnet

import (
	"fmt"
	"github.com/yddeng/utils/buffer"
	"io"
	"net"
	"reflect"
)

type defWsCodec struct {
	readBuf *buffer.Buffer
}

func newWsCodec() *defWsCodec {
	return &defWsCodec{readBuf: buffer.NewBufferWithCap(1024)}
}

//解码
func (decoder *defWsCodec) Decode(reader io.Reader) (interface{}, error) {
	n, err := decoder.readBuf.ReadFrom(reader)
	if err != nil {
		return nil, err
	}
	return decoder.readBuf.ReadBytes(int(n))
}

//编码
func (encoder *defWsCodec) Encode(o interface{}) ([]byte, error) {
	data, ok := o.([]byte)
	if !ok {
		return nil, fmt.Errorf("dnet:defWSCodec encode interface{} is %s, need type []byte", reflect.TypeOf(o))
	}
	return data, nil
}

type WSSession struct {
	*session
}

// NewWSSession return an initialized *WSSession
func NewWSSession(conn net.Conn, options ...Option) *WSSession {
	op := loadOptions(options...)
	if op.MsgCallback == nil {
		// need message callback
		panic(ErrNilMsgCallBack)
	}
	// init default codec
	if op.Codec == nil {
		op.Codec = newWsCodec()
	}

	return &WSSession{
		session: newSession(conn, op),
	}
}
