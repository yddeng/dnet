package dnet

import (
	"errors"
	"net"
	"os"
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

func NetError(err error) (brokenPipe bool) {
	if ne, ok := err.(*net.OpError); ok {
		if se, ok := ne.Err.(*os.SyscallError); ok {
			if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
				brokenPipe = true
			}
		}
	}
	return
}
