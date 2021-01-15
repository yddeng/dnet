package drpc

import (
	"errors"
	"fmt"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dutil/timer"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultRPCTimeout = 8 * time.Second
)

type Call struct {
	reqNo    uint64
	callback func(interface{}, error) // response
	timer    timer.Timer
}

type Client struct {
	reqNo    uint64         // 请求号
	timerMgr timer.TimerMgr // 定时器
	pending  sync.Map       // map[uint64]*Call
}

/*
 asynchronous 异步请求
*/
func (client *Client) Call(channel RPCChannel, method string, data interface{}, timeout time.Duration, callback func(interface{}, error)) error {
	if callback == nil {
		return errors.New("drpc: AsyncCall callback == nil")
	}

	req := &Request{
		SeqNo:  atomic.AddUint64(&client.reqNo, 1),
		Method: method,
		Data:   data,
	}

	if err := channel.SendRequest(req); err != nil {
		return err
	}

	c := &Call{
		reqNo:    req.SeqNo,
		callback: callback,
	}

	c.timer = client.timerMgr.OnceTimer(timeout, nil, func(ctx interface{}) {
		if _, ok := client.pending.Load(c.reqNo); ok {
			client.pending.Delete(c.reqNo)
			c.callback(nil, dnet.ErrRPCTimeout)
		}
		c.timer = nil
	})

	client.pending.Store(c.reqNo, c)
	return nil
}

func (client *Client) OnRPCResponse(resp *Response) error {
	v, ok := client.pending.Load(resp.SeqNo)
	if !ok {
		return fmt.Errorf("drpc: OnRPCResponse reqNo:%d is not found", resp.SeqNo)
	}

	call := v.(*Call)
	call.callback(resp.Data, nil)

	if call.timer != nil {
		call.timer.Stop()
	}
	client.pending.Delete(resp.SeqNo)
	return nil

}

// 默认低精度定时器
func NewClient() *Client {
	client := &Client{
		// 低精度定时器，精度50ms，长度20。误差 50ms
		timerMgr: timer.NewTimeWheelMgr(time.Millisecond*50, 20),
	}
	return client
}

func NewClientWithTimerMgr(timerMgr timer.TimerMgr) *Client {
	return &Client{
		timerMgr: timerMgr,
	}
}
