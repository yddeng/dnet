package dnet

import "fmt"

var (
	ErrNewClientNil = fmt.Errorf("newClient is nil")

	ErrStateFailed   = fmt.Errorf("session state failed")
	ErrNoCodec       = fmt.Errorf("session without codec")
	ErrNoMsgCallBack = fmt.Errorf("session without msgcallback")
	ErrSendMsgNil    = fmt.Errorf("session send msg is nil")
	ErrSendChanFull  = fmt.Errorf("session send chan is full")
)

var (
	ErrRPCTimeout = fmt.Errorf("rpc timeout")
)
