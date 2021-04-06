package dnet

import (
	"fmt"
	"github.com/yddeng/dutil/buffer"
	"reflect"
)

type (
	// 编码器
	Encoder interface {
		Encode(o interface{}) ([]byte, error)
	}

	// 解码器
	Decoder interface {
		Decode(b []byte) (interface{}, error)
	}

	//编解码器
	Codec interface {
		Encoder
		Decoder
	}
)

// default编解码器
// 消息 -- 格式: 消息头(消息len), 消息体

const (
	lenSize  = 2       // 消息长度（消息体的长度）
	headSize = lenSize // 消息头长度
	buffSize = 65535   // 缓存容量(与lenSize有关，2字节最大65535）
)

type defCodec struct {
	readBuf  *buffer.Buffer
	dataLen  uint16
	readHead bool
}

func NewCodec() *defCodec {
	return &defCodec{
		readBuf: &buffer.Buffer{},
	}
}

//解码
func (decoder *defCodec) Decode(b []byte) (interface{}, error) {
	_, _ = decoder.readBuf.Write(b)

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
func (encoder *defCodec) Encode(o interface{}) ([]byte, error) {

	data, ok := o.([]byte)
	if !ok {
		return nil, fmt.Errorf("encode interface{} is %s, need type []byte", reflect.TypeOf(o))
	}

	dataLen := len(data)
	if dataLen > buffSize-headSize {
		return nil, fmt.Errorf("encode dataLen is too large,len: %d", dataLen)
	}

	msgLen := dataLen + headSize
	buff := buffer.NewBufferWithCap(msgLen)

	//写入data长度
	buff.WriteUint16BE(uint16(dataLen))
	//data数据
	buff.WriteBytes(data)

	return buff.Bytes(), nil
}
