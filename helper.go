package dnet

import (
	"errors"
	"strings"
)

// 作为通知用的 channel， make(chan struct{}, 1)
func SendNotifyChan(ch chan struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}

func ParseAddr(addr string) (ip string, port string, err error) {
	idx := strings.Index(addr, ":")
	if idx == -1 {
		return "", "", errors.New("addr is failed")
	}
	return addr[:idx], addr[idx+1:], nil
}
