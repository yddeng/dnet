package dnet

import (
	"errors"
)

var (
	ErrNewClientNil = errors.New("dnet: newClient is nil")

	ErrStateFailed    = errors.New("dnet: session state failed")
	ErrNilCodec       = errors.New("dnet: session without codec")
	ErrNilMsgCallBack = errors.New("dnet: session without msgcallback")
	ErrSendMsgNil     = errors.New("dnet: session send msg is nil")
	ErrSendChanFull   = errors.New("dnet: session send chan is full")
	ErrSendTypeFailed = errors.New("dnet: session send message type is failed. only type([]byte) if encoder is nil. ")
)

var (
	ErrRPCTimeout = errors.New("dnet: rpc timeout")
)
