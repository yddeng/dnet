package dnet

import (
	"net"
)

type TCPSession struct {
	*session
	conn *net.TCPConn
}

// NewTCPSession return an initialized *TCPSession
func NewTCPSession(conn *net.TCPConn, options ...Option) (*TCPSession, error) {
	op := loadOptions(options...)
	if op.MsgCallback == nil {
		return nil, ErrNilMsgCallBack
	}
	// init default codec
	if op.Codec == nil {
		op.Codec = newTCPCodec()
	}

	tcpConn := &TCPSession{
		conn:    conn,
		session: newSession(conn, op),
	}
	return tcpConn, nil
}

//func (this *TCPSession) CloseRead() error {
//	return this.conn.CloseRead()
//}
//
//func (this *TCPSession) CloseWrite() error {
//	return this.conn.CloseWrite()
//}
