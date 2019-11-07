package rpc

import (
	"errors"
	"fmt"
	"github.com/yddeng/dutil/timingwheel"
	"sync"
	"sync/atomic"
	"time"
)

type Call struct {
	reqNo    uint64
	callback func(interface{}, error) //response
	deadline time.Time
	timer    *timingwheel.Timer
}

var ErrTimeout = errors.New("RPC timeout")

type Client struct {
	clientCodec ClientCodec
	reqNo       uint64   //请求号
	pending     sync.Map //map[uint64]*Call
	wheel       *timingwheel.TimingWheel
	timeout     time.Duration
}

/*
 异步请求
*/
func (client *Client) AsynCall(channel RPCChannel, req interface{}, callback func(interface{}, error)) error {
	if callback == nil {
		return errors.New("callback == nil")
	}

	msg := &Request{
		SeqNo:    atomic.AddUint64(&client.reqNo, 1),
		Data:     req,
		NeedResp: true,
	}

	data, err := client.clientCodec.EncodeRequest(msg)
	if err != nil {
		return err
	}
	err = channel.SendRequest(data)
	if err != nil {
		return err
	}

	c := &Call{
		reqNo:    msg.SeqNo,
		callback: callback,
		deadline: time.Now().Add(client.timeout),
	}

	c.timer = client.wheel.DelayFunc(client.timeout, func() {
		//fmt.Println("timeout -- ", c.reqNo)
		if _, ok := client.pending.Load(c.reqNo); ok {
			client.pending.Delete(c.reqNo)
			c.callback(nil, ErrTimeout)
		}
	})

	client.pending.Store(c.reqNo, c)
	return nil
}

/*
 同步请求
*/
func (client *Client) SynsCall(channel RPCChannel, req interface{}) (ret interface{}, err error) {
	sysnChan := make(chan struct{})
	err = client.AsynCall(channel, req, func(ret_ interface{}, err_ error) {
		ret = ret_
		err = err_
		sysnChan <- struct{}{}
	})
	if err == nil {
		_ = <-sysnChan
	}

	return
}

//只管将消息发送出去
func (client *Client) Post(channel RPCChannel, req interface{}) error {
	msg := &Request{
		SeqNo:    atomic.AddUint64(&client.reqNo, 1),
		Data:     req,
		NeedResp: false,
	}

	data, err := client.clientCodec.EncodeRequest(msg)
	if err != nil {
		return err
	}
	return channel.SendRequest(data)
}

func (client *Client) OnRPCResponse(date interface{}) error {
	resp, err := client.clientCodec.DecodeResponse(date)
	if err != nil {
		return err
	}

	v, ok := client.pending.Load(resp.SeqNo)
	if !ok {
		return fmt.Errorf("rpc call reqNo:%d not found\n", resp.SeqNo)
	}

	call := v.(*Call)
	call.callback(resp.Data, resp.Err)

	if call.timer != nil {
		//fmt.Println("delete timer", call.reqNo)
		client.wheel.RemoveTimer(call.timer)
	}
	client.pending.Delete(resp.SeqNo)
	return nil

}

func NewClient(clientCodec ClientCodec) *Client {
	client := &Client{
		clientCodec: clientCodec,
		timeout:     8 * time.Second, //rpc超时时间
	}
	client.wheel, _ = timingwheel.NewTimingWheel(10*time.Millisecond, 1000) //10s

	return client
}
