package dnet

import (
	"fmt"
	"github.com/yddeng/dutil/buffer"
	"io"
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
