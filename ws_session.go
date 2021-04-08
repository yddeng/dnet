package dnet

import (
	"github.com/gorilla/websocket"
)

type WSSession struct {
	*session
	conn *WSConn
}

func NewWSSession(conn *websocket.Conn, options ...Option) (*WSSession, error) {
	op := loadOptions(options...)
	if op.MsgCallback == nil {
		return nil, ErrNilMsgCallBack
	}
	if op.Codec == nil {
		op.Codec = newWsCodec()
	}

	wsConn := NewWSConn(conn)

	session := &WSSession{
		conn:    wsConn,
		session: newSession(wsConn, op),
	}

	return session, nil
}
