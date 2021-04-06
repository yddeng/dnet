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

	// 读buffer大小
	ReadBufferSize int

	// 读超时
	ReadTimeout time.Duration

	// 写超时
	WriteTimeout time.Duration

	// 消息回掉
	MsgCallback func(message interface{}, err error)

	// 编码器
	Encoder Encoder

	// 解码器
	Decoder Decoder

	// 关闭连接回调
	CloseCallback func(session Session, reason error)
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

func WithReadBufferSize(size int) Option {
	return func(opt *Options) {
		opt.ReadBufferSize = size
	}
}

func WithSendChannelSize(size int) Option {
	return func(opt *Options) {
		opt.SendChannelSize = size
	}
}

func WithMessageCallback(msgCb func(message interface{}, err error)) Option {
	return func(opt *Options) {
		opt.MsgCallback = msgCb
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
		opt.Encoder = codec
		opt.Decoder = codec
	}
}

func WithEncoder(encoder Encoder) Option {
	return func(opt *Options) {
		opt.Encoder = encoder
	}
}

func WithDecoder(decoder Decoder) Option {
	return func(opt *Options) {
		opt.Decoder = decoder
	}
}

func WithCloseCallback(closeCallback func(session Session, reason error)) Option {
	return func(opt *Options) {
		opt.CloseCallback = closeCallback
	}
}
