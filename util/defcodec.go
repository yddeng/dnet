package util

import (
	"fmt"
	"github.com/yddeng/dutil/buffer"
	"io"
	"reflect"
)

// default编解码器
// 消息 -- 格式: 消息头(消息len), 消息体

const (
	lenSize  = 2       // 消息长度（消息体的长度）
	headSize = lenSize // 消息头长度
	buffSize = 65535   // 缓存容量(与lenSize有关，2字节最大65535）
)

type Codec struct {
	readBuf *buffer.Buffer
	dataLen uint16
}

func NewCodec() *Codec {
	return &Codec{
		readBuf: &buffer.Buffer{},
	}
}

//解码
func (decoder *Codec) Decode(reader io.Reader) (interface{}, error) {
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

func (decoder *Codec) unPack() ([]byte, error) {

	if decoder.dataLen == 0 {
		if decoder.readBuf.Len() < headSize {
			return nil, nil
		}

		decoder.dataLen, _ = decoder.readBuf.ReadUint16BE()
	}

	if decoder.readBuf.Len() < int(decoder.dataLen) {
		return nil, nil
	}

	data, _ := decoder.readBuf.ReadBytes(int(decoder.dataLen))

	//将消息长度置为0，用于下一次验证
	decoder.dataLen = 0
	return data, nil
}

//编码
func (encoder *Codec) Encode(o interface{}) ([]byte, error) {

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
