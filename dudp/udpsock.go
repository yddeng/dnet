package dudp

import (
	"github.com/xtaci/kcp-go"
	"github.com/yddeng/dnet"
	"net"
	"sync"
	"time"
)

const (
	started = 0x01 //0000 0001
	closed  = 0x02 //0000 0010
)

const sendBufChanSize = 1024

type UDPConn struct {
	state        byte
	conn         *kcp.UDPSession
	ctx          interface{}   // 用户数据
	readTimeout  time.Duration // 读超时
	writeTimeout time.Duration // 写超时

	codec dnet.Codec // 编解码器

	sendNotifyCh  chan struct{} // 发送消息通知
	sendMessageCh chan *message // 发送队列

	msgCallback func(interface{}, error) // 消息回调

	closeCallback func(session dnet.Session, reason string) // 关闭连接回调
	closeReason   string                                    // 关闭原因

	lock sync.Mutex
}
