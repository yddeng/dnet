// +build darwin netbsd freebsd openbsd dragonfly

package poller

import (
	"syscall"
)

type Poller struct {
	fd     int
	events []syscall.Kevent_t
}

func OpenPoller() (poller *Poller, err error) {
	poller = new(Poller)
	poller.events = make([]syscall.Kevent_t, waitEventsBegin)
	if poller.fd, err = syscall.Kqueue(); err != nil {
		poller = nil
		return
	}
	if _, err = syscall.Kevent(poller.fd, []syscall.Kevent_t{{
		Ident:  0,
		Filter: syscall.EVFILT_USER,
		Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
	}}, nil, nil); err != nil {
		_ = poller.Close()
		poller = nil
		return
	}
	return
}

func (p *Poller) Close() error {
	return syscall.Close(p.fd)
}

var wakeChanges = []syscall.Kevent_t{{
	Ident:  0,
	Filter: syscall.EVFILT_USER,
	Fflags: syscall.NOTE_TRIGGER,
}}

func (p *Poller) Trigger() error {
	_, err := syscall.Kevent(p.fd, wakeChanges, nil, nil)
	return err
}

// AddReadWrite 注册fd到kqueue并注册可读可写事件
func (p *Poller) AddReadWrite(fd int) error {
	_, err := syscall.Kevent(p.fd, []syscall.Kevent_t{
		{Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_READ},
		{Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_WRITE},
	}, nil, nil)
	return err

}

// AddRead 注册fd到kqueue并注册可读事件
func (p *Poller) AddRead(fd int) error {
	_, err := syscall.Kevent(p.fd, []syscall.Kevent_t{
		{Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_READ}}, nil, nil)
	return err

}

// AddWrite 注册fd到kqueue并注册可写事件
func (p *Poller) AddWrite(fd int) error {
	_, err := syscall.Kevent(p.fd, []syscall.Kevent_t{
		{Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_WRITE}}, nil, nil)
	return err
}

// ModRead 修改fd注册事件为可读事件
func (p *Poller) ModRead(fd int) error {
	_, err := syscall.Kevent(p.fd, []syscall.Kevent_t{
		{Ident: uint64(fd), Flags: syscall.EV_DELETE, Filter: syscall.EVFILT_WRITE}}, nil, nil)
	return err
}

// ModReadWrite 修改fd注册事件为可读可写事件
func (p *Poller) ModReadWrite(fd int) error {
	_, err := syscall.Kevent(p.fd, []syscall.Kevent_t{
		{Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_WRITE}}, nil, nil)
	return err
}

// Del 从kqueue删除fd
func (p *Poller) Delete(fd int) error {
	_, err := syscall.Kevent(p.fd, []syscall.Kevent_t{
		{Ident: uint64(fd), Flags: syscall.EV_DELETE, Filter: syscall.EVFILT_WRITE},
		{Ident: uint64(fd), Flags: syscall.EV_DELETE, Filter: syscall.EVFILT_READ}}, nil, nil)
	return err
}

func (p *Poller) Polling(callback func(fd int, ev Event) error) (err error) {
	var wakeUp bool
	for {
		n, err := syscall.Kevent(p.fd, nil, p.events, nil)
		if err != nil && err != syscall.EINTR {
			return err
		}

		var event Event
		var e syscall.Kevent_t
		for i := 0; i < n; i++ {
			if fd := int(p.events[i].Ident); fd != 0 {
				event = 0
				e = p.events[i]
				if (e.Flags&syscall.EV_EOF != 0) || (e.Flags&syscall.EV_ERROR != 0) {
					event |= EventErr
				}
				if e.Filter == syscall.EVFILT_READ {
					event |= EventRead
				}
				if e.Filter == syscall.EVFILT_WRITE {
					event |= EventWrite
				}
				if err = callback(fd, event); err != nil {
					//return
				}
			} else {
				wakeUp = true
			}
		}
		if wakeUp {
			if err = callback(-1, EventNone); err != nil {
				//return
			}
			wakeUp = false
		}
		if n == len(p.events) {
			p.events = make([]syscall.Kevent_t, n<<1)
		}
	}
}
