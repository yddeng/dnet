package codec

import (
	"github.com/tagDong/dnet"
	"github.com/tagDong/dnet/module/message"
	"github.com/tagDong/dnet/module/protocol"
	"github.com/tagDong/dnet/util"
	"net"
)

// 编解码器
// 消息 -- 格式: 消息头(消息len＋消息cmd+消息ID), 消息体

const (
	lenSize  = 2                          // 消息长度（消息体的长度）
	cmdSize  = 2                          // 消息规则（目前为消息的索引）
	idSize   = 2                          // 消息ID（消息体的编码ID，对应的反序列化结构）
	headSize = lenSize + cmdSize + idSize // 消息头长度
	buffSize = 65535                      // 缓存容量(与lenSize有关，2字节最大65535）
)

type Codec struct {
	*Decoder
	procol *protocol.Protocol
}

func NewCodec(procol *protocol.Protocol) *Codec {
	return &Codec{
		Decoder: &Decoder{readBuf: util.NewBuffer(buffSize)},
		procol:  procol,
	}
}

type Decoder struct {
	readBuf *util.Buffer
	msgLen  uint16
	cmd     uint16
	msgID   uint16
}

//解码
func (decoder *Codec) Decode(conn net.Conn) (dnet.Message, error) {
	for {
		msg, err := decoder.unPack()

		//fmt.Println(msg, err)
		if msg != nil {
			return msg, nil

		} else if err == nil {
			_, err1 := decoder.readBuf.ReadFrom(conn)
			if err1 != nil {
				return nil, err1
			}
		} else {
			return nil, err
		}
	}
}

func (decoder *Codec) unPack() (dnet.Message, error) {

	if decoder.msgLen == 0 {
		if decoder.readBuf.Len() < headSize {
			return nil, nil
		}

		decoder.msgLen, _ = decoder.readBuf.ReadUint16BE()
		decoder.cmd, _ = decoder.readBuf.ReadUint16BE()
		decoder.msgID, _ = decoder.readBuf.ReadUint16BE()

	}

	if decoder.readBuf.Len() < int(decoder.msgLen) {
		return nil, nil
	}

	data, _ := decoder.readBuf.ReadBytes(int(decoder.msgLen))

	msg, err := decoder.procol.Unmarshal(decoder.msgID, data)
	if err != nil {
		return nil, err
	}

	seriNo := decoder.cmd
	ret := message.NewMessage(seriNo, msg)

	//将消息长度置为0，用于下一次验证
	decoder.msgLen = 0
	return ret, nil
}

//编码
func (encoder *Codec) Encode(msg dnet.Message) ([]byte, error) {

	msgID, data, err := encoder.procol.Marshal(msg.GetData())
	if err != nil {
		return nil, err
	}

	msgLen := len(data) + headSize

	buff := util.NewBuffer(msgLen)

	//msgLen+cmd+msgID
	//写入data长度
	buff.AppendUint16BE(uint16(len(data)))
	//写入cmd
	buff.AppendUint16BE(msg.GetSerialNo())
	//msgID
	buff.AppendUint16BE(msgID)
	//data数据
	buff.AppendBytes(data)

	//fmt.Println("encode", len(data), msgID, msg.GetSerialNo(), data, buff.Peek(), buff.Len())

	return buff.Peek(), nil
}
