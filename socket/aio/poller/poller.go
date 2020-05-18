package poller

const waitEventsBegin = 64

type Event uint32

// Event poller 返回事件值
const (
	EventRead  Event = 0x1
	EventWrite Event = 0x2
	EventErr   Event = 0x80
	EventNone  Event = 0
)
