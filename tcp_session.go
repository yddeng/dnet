package dnet

type TCPSession struct {
	*session
}

// NewTCPSession return an initialized *TCPSession
func NewTCPSession(conn NetConn, options ...Option) (*TCPSession, error) {
	op := loadOptions(options...)
	if op.MsgCallback == nil {
		return nil, ErrNilMsgCallBack
	}
	// init default codec
	if op.Codec == nil {
		op.Codec = newTCPCodec()
	}

	tcpConn := &TCPSession{
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
