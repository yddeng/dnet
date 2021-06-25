// +build linux

package poller

import (
	"syscall"
)

type Poller struct {
	fd     int
	wfd    int
	wfdBuf []byte
	events []syscall.EpollEvent
}

func OpenPoller() (poller *Poller, err error) {
	poller = new(Poller)
	if poller.fd, err = syscall.EpollCreate1(0); err != nil {
		poller = nil
		return
	}
	r0, _, e1 := syscall.Syscall(syscall.SYS_EVENTFD2, 0 /*syscall.O_NONBLOCK*/, 0, 0)
	if e1 != 0 {
		_ = syscall.Close(poller.fd)
		poller = nil
		return
	}

	poller.wfd = int(r0)
	poller.wfdBuf = make([]byte, 8)
	poller.events = make([]syscall.EpollEvent, waitEventsBegin)

	if err = poller.AddRead(poller.wfd); err != nil {
		_ = poller.Close()
		poller = nil
		return
	}

	return
}

func (p *Poller) Close() error {
	if err := syscall.Close(p.fd); err != nil {
		return err
	}
	return syscall.Close(p.wfd)
}

var wakeBytes = []byte{0, 0, 0, 0, 0, 0, 0, 1}

func (p *Poller) Trigger() error {
	_, err := syscall.Write(p.wfd, wakeBytes)
	return err
}

const (
	readEvents      = syscall.EPOLLPRI | syscall.EPOLLIN
	writeEvents     = syscall.EPOLLOUT
	readWriteEvents = readEvents | writeEvents
	errorEvents     = int(syscall.EPOLLERR | syscall.EPOLLHUP | syscall.EPOLLRDHUP)
)

// AddReadWrite 注册fd到 epoll，并注册可读可写事件
func (p *Poller) AddReadWrite(fd int) error {
	return syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{Fd: int32(fd), Events: readWriteEvents})
}

// AddRead 注册fd到 epoll，并注册可读事件
func (p *Poller) AddRead(fd int) error {
	return syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{Fd: int32(fd), Events: readEvents})
}

// AddWrite 注册fd到 epoll，并注册可写事件
func (p *Poller) AddWrite(fd int) error {
	return syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{Fd: int32(fd), Events: writeEvents})
}

// ModRead 修改fd注册事件为可读事件
func (p *Poller) ModRead(fd int) error {
	return syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_MOD, fd, &syscall.EpollEvent{Fd: int32(fd), Events: readEvents})
}

// ModWrite 修改fd注册事件为可读事件
func (p *Poller) ModWrite(fd int) error {
	return syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_MOD, fd, &syscall.EpollEvent{Fd: int32(fd), Events: writeEvents})
}

// ModReadWrite 修改fd注册事件为可读可写事件
func (p *Poller) ModReadWrite(fd int) error {
	return syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_MOD, fd, &syscall.EpollEvent{Fd: int32(fd), Events: readWriteEvents})
}

// Del 从 epoll中删除fd
func (p *Poller) Delete(fd int) error {
	return syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_DEL, fd, nil)
}

// Poll 启动 epoll wait 循环
func (p *Poller) Polling(callback func(fd int, ev Event) error) (err error) {
	var wakeUp bool
	var event Event
	var e syscall.EpollEvent
	for {
		n, err := syscall.EpollWait(p.fd, p.events, -1)
		if err != nil && err != syscall.EINTR {
			return err
		}
		for i := 0; i < n; i++ {
			e = p.events[i]
			if fd := int(e.Fd); fd != p.wfd {
				if e.Events&uint32(errorEvents) != 0 {
					event |= EventErr
				}
				if e.Events&uint32(readEvents) != 0 {
					event |= EventRead
				}
				if e.Events&uint32(writeEvents) != 0 {
					event |= EventWrite
				}
				if err = callback(fd, event); err != nil {
					return
				}
			} else {
				wakeUp = true
				_, _ = syscall.Read(p.wfd, p.wfdBuf)
			}
		}
		if wakeUp {
			if err = callback(-1, EventNone); err != nil {
				return
			}
			wakeUp = false
		}
		if n == len(p.events) {
			p.events = make([]syscall.EpollEvent, n<<1)
		}
	}
}
