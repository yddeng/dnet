package protobuf

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"reflect"
)

type PbProtocol struct {
	id2type map[uint16]reflect.Type
	type2id map[reflect.Type]uint16
}

func newPbProtocol() *PbProtocol {
	return &PbProtocol{
		id2type: map[uint16]reflect.Type{},
		type2id: map[reflect.Type]uint16{},
	}
}

var pbuf = newPbProtocol()

func Register(id uint16, pb proto.Message) {

	tt := reflect.TypeOf(pb)

	if _, ok := pbuf.id2type[id]; ok {
		fmt.Errorf("%d already register to type:%s", id, tt.String())
		return
	}

	pbuf.id2type[id] = tt
	pbuf.type2id[tt] = id
}

func Marshal(data interface{}) (uint16, []byte, error) {
	id, ok := pbuf.type2id[reflect.TypeOf(data)]
	if !ok {
		return 0, nil, fmt.Errorf("type: %s undefined", reflect.TypeOf(data))
	}

	ret, err := proto.Marshal(data.(proto.Message))
	if err != nil {
		return 0, nil, err
	}

	return id, ret, nil
}

func Unmarshal(msgID uint16, data []byte) (msg proto.Message, err error) {
	tt, ok := pbuf.id2type[msgID]

	if !ok {
		err = fmt.Errorf("msgID: %d undefined", msgID)
		return
	}

	msg = reflect.New(tt.Elem()).Interface().(proto.Message)

	err = proto.Unmarshal(data, msg)

	if err != nil {
		return nil, err
	}

	return msg, nil
}
