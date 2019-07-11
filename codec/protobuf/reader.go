package protobuf

import (
	"github.com/golang/protobuf/proto"
	"github.com/tagDong/dnet/util"
	"net"
)

// 消息 -- 格式: 消息头(消息大小＋消息id), 消息体
const (
	lenSize  = 2                // 消息大小所占空间(header1)
	idSize   = 2                // id大小(header2)
	headSize = lenSize + idSize // 消息头(消息大小＋消息id)大小
	buffSize = 65535            // 缓存容量
)

type Reader struct {
	buff *util.Buffer
}

func NewReader() *Reader {
	return &Reader{
		buff: util.NewBuffer(buffSize),
	}
}

func (r *Reader) Receive(conn net.Conn) (interface{}, error) {
	for {
		msg, err := r.unPack()

		//fmt.Println(msg, err)
		if msg != nil {
			return msg, nil

		} else if err == nil {
			_, err1 := r.buff.ReadFrom(conn)
			if err1 != nil {
				return nil, err1
			}
		} else {
			return nil, err
		}
	}
}

func (r *Reader) unPack() (interface{}, error) {

	if r.buff.Len() < headSize {
		return nil, nil
	} else {

		var msgLen, msgID uint16

		var msg proto.Message
		var err error

		msgLen = util.GetUint16BE(r.buff.ReadBytes(0, lenSize))
		msgID = util.GetUint16BE(r.buff.ReadBytes(lenSize, idSize))

		//fmt.Println(msgLen, msgID, r.buff.Len())
		if lenSize+int(msgLen) > r.buff.Len() {
			return nil, nil

		}

		var msgData = make([]byte, int(msgLen-idSize))
		copy(msgData, r.buff.ReadBytes(headSize, int(msgLen-idSize)))

		//fmt.Println(msgLen, msgID, msgData)
		msg, err = Unmarshal(msgID, msgData)
		if err != nil {
			return nil, err
		}

		r.buff.Reset(int(msgLen + lenSize))
		return msg, nil

	}
}
