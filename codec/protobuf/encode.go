package protobuf

import (
	"github.com/tagDong/dnet/util"
)

type Encode struct{}

func NewEncode() *Encode {
	return &Encode{}
}

func (w *Encode) Pack(msg interface{}) ([]byte, error) {
	id, data, err := Marshal(msg)
	if err != nil {
		return nil, err
	}

	msgLen := len(data) + headSize

	buff := util.NewBuffer(msgLen)

	buff.AppendUint16(uint16(msgLen - lenSize))

	buff.AppendUint16(id)

	buff.AppendBytes(data)

	//fmt.Println("ss", id, data, buff.Bytes(), buff.Len())

	return buff.Bytes(), nil
}
