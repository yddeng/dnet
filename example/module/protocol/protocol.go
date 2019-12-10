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

var pProtocol *Protocol

func InitProtocol(protoc Protoc) {
	pProtocol = &Protocol{
		id2Type: map[uint16]reflect.Type{},
		type2Id: map[reflect.Type]uint16{},
		protoc:  protoc,
	}
}

func Register(id uint16, msg interface{}) {
	if pProtocol == nil {
		fmt.Errorf("protocol is nil,need init")
		return
	}

	tt := reflect.TypeOf(msg)

	if _, ok := pProtocol.id2Type[id]; ok {
		fmt.Errorf("%d already register to type:%s", id, tt)
		return
	}

	pProtocol.id2Type[id] = tt
	pProtocol.type2Id[tt] = id
}

func Marshal(data interface{}) (uint16, []byte, error) {
	if pProtocol == nil {
		return 0, nil, fmt.Errorf("protocol is nil,need init")
	}

	id, ok := pProtocol.type2Id[reflect.TypeOf(data)]
	if !ok {
		return 0, nil, fmt.Errorf("type: %s undefined", reflect.TypeOf(data))
	}

	ret, err := pProtocol.protoc.Marshaler(data)
	if err != nil {
		return 0, nil, err
	}

	return id, ret, nil
}

func Unmarshal(msgID uint16, data []byte) (msg interface{}, err error) {
	if pProtocol == nil {
		return nil, fmt.Errorf("protocol is nil,need init")
	}

	tt, ok := pProtocol.id2Type[msgID]
	if !ok {
		err = fmt.Errorf("msgID: %d undefined", msgID)
		return
	}

	//反序列化的结构
	msg = reflect.New(tt.Elem()).Interface()
	err = pProtocol.protoc.Unmarshaler(data, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}
