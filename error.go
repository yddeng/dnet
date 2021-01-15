package dnet

import "fmt"

var (
	ErrNewClientNil = fmt.Errorf("dnet: newClient is nil")

	ErrStateFailed   = fmt.Errorf("dnet: session state failed")
	ErrNoCodec       = fmt.Errorf("dnet: session without codec")
	ErrNoMsgCallBack = fmt.Errorf("dnet: session without msgcallback")
	ErrSendMsgNil    = fmt.Errorf("dnet: session send msg is nil")
	ErrSendChanFull  = fmt.Errorf("dnet: session send chan is full")
)

var (
	ErrRPCTimeout = fmt.Errorf("dnet: rpc timeout")
)
