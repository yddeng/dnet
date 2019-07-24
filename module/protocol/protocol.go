package protocol

import (
	"fmt"
	"reflect"
)

//协议
type Protoc interface {
	//反序列化
	Unmarshaler(data []byte, o interface{}) (err error)
	//序列化
	Marshaler(data interface{}) ([]byte, error)
}

type Protocol struct {
	id2Type map[uint16]reflect.Type
	type2Id map[reflect.Type]uint16
	protoc  Protoc
}

func NewProtocol(protoc Protoc) *Protocol {
	return &Protocol{
		id2Type: map[uint16]reflect.Type{},
		type2Id: map[reflect.Type]uint16{},
		protoc:  protoc,
	}
}

func (pb *Protocol) RegisterIDMsg(id uint16, msg interface{}) {
	tt := reflect.TypeOf(msg)

	if _, ok := pb.id2Type[id]; ok {
		fmt.Errorf("%d already register to type:%s", id, tt)
		return
	}

	pb.id2Type[id] = tt
	pb.type2Id[tt] = id
}

func (pb *Protocol) Marshal(data interface{}) (uint16, []byte, error) {
	id, ok := pb.type2Id[reflect.TypeOf(data)]
	if !ok {
		return 0, nil, fmt.Errorf("type: %s undefined", reflect.TypeOf(data))
	}

	ret, err := pb.protoc.Marshaler(data)
	if err != nil {
		return 0, nil, err
	}

	return id, ret, nil
}

func (pb *Protocol) Unmarshal(msgID uint16, data []byte) (msg interface{}, err error) {
	tt, ok := pb.id2Type[msgID]
	if !ok {
		err = fmt.Errorf("msgID: %d undefined", msgID)
		return
	}

	//反序列化的结构
	msg = reflect.New(tt.Elem()).Interface()
	err = pb.protoc.Unmarshaler(data, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}
