package dnet

type WSSession struct {
	*session
}

// NewWSSession return an initialized *WSSession
func NewWSSession(conn NetConn, options ...Option) (*WSSession, error) {
	op := loadOptions(options...)
	if op.MsgCallback == nil {
		return nil, ErrNilMsgCallBack
	}
	// init default codec
	if op.Codec == nil {
		op.Codec = newWsCodec()
	}

	session := &WSSession{
		session: newSession(conn, op),
	}

	return session, nil
}
