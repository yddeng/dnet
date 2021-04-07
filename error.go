package dnet

import (
	"errors"
)

var (
	ErrSessionClosed  = errors.New("dnet: session is closed. ")
	ErrNilMsgCallBack = errors.New("dnet: session without msgCallback")
	ErrSendMsgNil     = errors.New("dnet: session send msg is nil")
	ErrSendChanFull   = errors.New("dnet: session send channel is full")

	ErrSendTimeout = errors.New("dnet: send timeout. ")
	ErrReadTimeout = errors.New("dnet: read timeout. ")
)

var (
	ErrRPCTimeout = errors.New("dnet: rpc timeout")
)
