package rpc

import (
	"errors"
	"fmt"
	"github.com/yddeng/dnet"
	"sync"
	"sync/atomic"
	"time"
)

type Call struct {
	reqNo    uint64
	callback func(interface{}, error) //response
	timer    *time.Timer
}

var timeout = 8 * time.Second

type Client struct {
	reqNo   uint64   //请求号
	pending sync.Map //map[uint64]*Call
}

/*
 异步请求
*/
func (client *Client) AsynCall(channel RPCChannel, msg interface{}, callback func(interface{}, error)) error {
	if callback == nil {
		return errors.New("callback == nil")
	}

	req := &Request{
		SeqNo:    atomic.AddUint64(&client.reqNo, 1),
		Data:     msg,
		NeedResp: true,
	}

	err := channel.SendRequest(req)
	if err != nil {
		return err
	}

	c := &Call{
		reqNo:    req.SeqNo,
		callback: callback,
	}

	c.timer = time.AfterFunc(timeout, func() {
		if _, ok := client.pending.Load(c.reqNo); ok {
			client.pending.Delete(c.reqNo)
			c.callback(nil, dnet.ErrRPCTimeout)
		}
	})

	client.pending.Store(c.reqNo, c)
	return nil
}

// 需用户在逻辑层实现。在同一线程处理会导致死锁。
/*
 同步请求
*/
//func (client *Client) SynsCall(channel RPCChannel, msg interface{}) (ret interface{}, err error) {
//	sysnChan := make(chan struct{})
//	err = client.AsynCall(channel, msg, func(ret_ interface{}, err_ error) {
//		ret = ret_
//		err = err_
//		sysnChan <- struct{}{}
//	})
//	if err == nil {
//		_ = <-sysnChan
//	}
//
//	return
//}

//只管将消息发送出去,
func (client *Client) Post(channel RPCChannel, msg interface{}) error {
	req := &Request{
		SeqNo:    atomic.AddUint64(&client.reqNo, 1),
		Data:     msg,
		NeedResp: false,
	}

	return channel.SendRequest(req)
}

func (client *Client) OnRPCResponse(resp *Response) error {
	v, ok := client.pending.Load(resp.SeqNo)
	if !ok {
		return fmt.Errorf("rpc call reqNo:%d not found", resp.SeqNo)
	}

	call := v.(*Call)
	call.callback(resp.Data, resp.Err)

	if call.timer != nil {
		call.timer.Stop()
	}
	client.pending.Delete(resp.SeqNo)
	return nil

}

func NewClient() *Client {
	client := &Client{}
	return client
}
