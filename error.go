package dnet

import "fmt"

var (
	ErrNewClientNil = fmt.Errorf("newClient is nil")

	ErrSessionStarted = fmt.Errorf("session is already started")
	ErrNotStarted     = fmt.Errorf("session is not started")
	ErrSessionClosed  = fmt.Errorf("session is closed")
	ErrNoCodec        = fmt.Errorf("session without codec")
	ErrNoMsgCallBack  = fmt.Errorf("session without msgcallback")
	ErrSendMsgNil     = fmt.Errorf("session send msg is nil")
	ErrSendChanFull   = fmt.Errorf("session send chan is full")
)
