package dnet

import "time"

type Option func(opt *Options)

func loadOptions(options ...Option) *Options {
	opts := new(Options)
	for _, option := range options {
		option(opts)
	}
	return opts
}

type Options struct {
	// 当写入队列满时，如果 BlockSend 为真，将阻塞。为假，返回队列满错误码. 默认为假
	BlockSend bool

	// 发送队列容量
	SendChannelSize int

	// 读超时
	ReadTimeout time.Duration

	// 写超时
	WriteTimeout time.Duration

	// 消息回掉
	MsgCallback func(session Session, message interface{})

	// 错误回掉
	ErrorCallback func(session Session, err error)

	// 关闭连接回调
	CloseCallback func(session Session, reason error)

	// 编解码器
	Codec Codec
}

func WithOptions(option *Options) Option {
	return func(opt *Options) {
		opt = option
	}
}

func WithBlockSend(bs bool) Option {
	return func(opt *Options) {
		opt.BlockSend = bs
	}
}

func WithSendChannelSize(size int) Option {
	return func(opt *Options) {
		opt.SendChannelSize = size
	}
}

func WithMessageCallback(msgCb func(session Session, message interface{})) Option {
	return func(opt *Options) {
		opt.MsgCallback = msgCb
	}
}

func WithErrorCallback(errCb func(session Session, err error)) Option {
	return func(opt *Options) {
		opt.ErrorCallback = errCb
	}
}

func WithTimeout(readTimeout, writeTimeout time.Duration) Option {
	return func(opt *Options) {
		opt.ReadTimeout = readTimeout
		opt.WriteTimeout = writeTimeout
	}
}

func WithCodec(codec Codec) Option {
	return func(opt *Options) {
		opt.Codec = codec
	}
}

func WithCloseCallback(closeCallback func(session Session, reason error)) Option {
	return func(opt *Options) {
		opt.CloseCallback = closeCallback
	}
}
