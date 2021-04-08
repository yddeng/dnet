package dnet

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type session struct {
	opts *Options

	conn    NetConn
	context atomic.Value //interface{} // 用户数据

	sendOnce      sync.Once
	sendNotifyCh  chan struct{} // 发送消息通知
	sendMessageCh chan *message // 发送队列

	closed    bool
	waitGroup sync.WaitGroup
	lock      sync.Mutex
}

type message struct {
	data interface{}
}

func newSession(conn NetConn, options *Options) *session {
	if options.SendChannelSize <= 0 {
		options.SendChannelSize = defSendChannelSize
	}

	session := &session{
		conn:         conn,
		opts:         options,
		sendNotifyCh: make(chan struct{}, 1),
	}

	go session.readThread()

	return session
}

func (this *session) SetContext(context interface{}) {
	this.context.Store(context)
}

func (this *session) Context() interface{} {
	return this.context.Load()
}

func (this *session) IsClosed() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.closed
}

func (this *session) NetConn() interface{} {
	return this.conn
}

func (this *session) LocalAddr() net.Addr {
	return this.conn.LocalAddr()
}

//对端地址
func (this *session) RemoteAddr() net.Addr {
	return this.conn.RemoteAddr()
}

// 接收线程
func (this *session) readThread() {
	this.waitGroup.Add(1)
	defer this.waitGroup.Done()

	for {

		if this.opts.ReadTimeout > 0 {
			if err := this.conn.SetReadDeadline(time.Now().Add(this.opts.ReadTimeout)); err != nil {
				if this.opts.ErrorCallback != nil {
					this.opts.ErrorCallback(this, err)
				}
			}
		}

		if msg, err := this.opts.Codec.Decode(this.conn); this.IsClosed() {
			break

		} else {
			if err != nil {
				if ne, ok := err.(net.Error); ok {
					if ne.Timeout() {
						err = ErrReadTimeout
					}
				}

				if this.opts.ErrorCallback != nil {
					this.opts.ErrorCallback(this, err)
				}
				this.Close(err)
				break

			} else if msg != nil {
				this.opts.MsgCallback(this, msg)
			}

		}

	}
}

// 发送线程
// 关闭连接时，发送完后再关闭
func (this *session) writeThread() {
	this.waitGroup.Add(1)
	defer this.waitGroup.Done()

	for {
		select {
		case msg := <-this.sendMessageCh:
			if this.IsClosed() {
				return
			}
			if data, err := this.opts.Codec.Encode(msg.data); err != nil {
				if !this.IsClosed() {
					if this.opts.ErrorCallback != nil {
						this.opts.ErrorCallback(this, err)
					}
					this.Close(err)
				}
				return
			} else {
				if data != nil && len(data) != 0 {
					// 发送的消息
					if this.opts.WriteTimeout > 0 {
						if err := this.conn.SetWriteDeadline(time.Now().Add(this.opts.ReadTimeout)); err != nil {
							if this.opts.ErrorCallback != nil {
								this.opts.ErrorCallback(this, err)
							}
						}
					}

					idx, length := 0, len(data)
					for idx < length {
						if n, err := this.conn.Write(data[idx:length]); err != nil {
							if !this.IsClosed() {
								if ne, ok := err.(net.Error); ok {
									if ne.Timeout() {
										err = ErrSendTimeout
									}
								}
								if this.opts.ErrorCallback != nil {
									this.opts.ErrorCallback(this, err)
								}
								this.Close(err)
							}
							return
						} else {
							idx += n
						}
					}
				}
			}

		default:
			if this.IsClosed() {
				return
			} else {
				// 等待发送事件
				<-this.sendNotifyCh
			}
		}

	}
}

func (this *session) Send(o interface{}) error {
	if o == nil {
		return ErrSendMsgNil
	}

	if this.IsClosed() {
		return ErrSessionClosed
	}

	if !this.opts.BlockSend {
		if len(this.sendMessageCh) == this.opts.SendChannelSize {
			return ErrSendChanFull
		}
	}

	this.sendOnce.Do(func() {
		this.sendMessageCh = make(chan *message, this.opts.SendChannelSize)
		go this.writeThread()
	})

	this.sendMessageCh <- &message{
		data: o,
	}
	sendNotifyChan(this.sendNotifyCh)

	return nil
}

/*
 主动关闭连接
 先关闭读，待写发送完毕关闭写
*/
func (this *session) Close(reason error) {
	this.lock.Lock()
	if this.closed {
		this.lock.Unlock()
		return
	}

	this.closed = true
	this.lock.Unlock()

	//_ = this.conn.(*net.TCPConn).CloseRead()
	_ = this.conn.Close()
	// 触发循环
	sendNotifyChan(this.sendNotifyCh)

	go func() {
		this.waitGroup.Wait()
		//_ = this.conn.Close()
		if this.opts.CloseCallback != nil {
			this.opts.CloseCallback(this, reason)
		}
	}()
}

// 作为通知用的 channel， make(chan struct{}, 1)
func sendNotifyChan(ch chan struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}
