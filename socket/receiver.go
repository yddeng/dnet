package socket

import (
	"github.com/tagDong/dnet/util"
	"net"
)

// 消息 -- 格式: 消息长度, 消息id, 消息体
// 这里解析出: 消息长度 + data（消息ID，消息体）。
// data的解析交给上层

// 常量
const (
	lenSize  = 4      //消息长度所占空间
	buffSize = 0xFFFF //65535 缓存区大小
)

type ReceiverI interface {
	ReadAndUnPack(reader net.Conn) ([]byte, error)
}

type Receiver struct {
	*util.Buffer
}

func NewReceiver() *Receiver {
	return &Receiver{
		Buffer: util.NewBuffer(buffSize),
	}
}

func (r *Receiver) ReadAndUnPack(conn net.Conn) ([]byte, error) {
	for {

		msg, err := r.unPack()

		if msg != nil {
			return msg, nil

		} else if err == nil {

			_, err := r.ReadFrom(conn)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

}

func (r *Receiver) unPack() ([]byte, error) {
	if r.Buffer.Size() < lenSize {
		return nil, nil
	} else {

		msgLen, err := r.Uint32(r.Buffer.Bytes(0, lenSize))
		if err != nil {
			return nil, err
		}

		if r.Buffer.Size()-lenSize < int(msgLen) {
			return nil, nil

		} else {
			var msg = make([]byte, int(msgLen))
			copy(msg, r.Buffer.Bytes(lenSize, int(msgLen)))
			r.Buffer.Reset(lenSize + int(msgLen))
			return msg, nil
		}
	}
}
