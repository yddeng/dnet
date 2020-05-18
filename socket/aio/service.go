package aio

import (
	"errors"
	"github.com/yddeng/dnet/socket/aio/poller"
	"log"
	"runtime"
	"sync"
	"syscall"
)

type Service struct {
	loops       []*EventLoop
	nextLoopIdx int
}

func NewService(loopCount int) *Service {
	if loopCount <= 0 {
		loopCount = runtime.NumCPU() * 2
	}

	s := &Service{loops: make([]*EventLoop, loopCount)}

	for i := 0; i < loopCount; i++ {
		loop, err := NewEventLoop()
		if err != nil {
			return nil
		}
		s.loops[i] = loop
	}

	return s
}

func (s *Service) NextEventLoop() *EventLoop {
	// todo 更多的负载方式
	loop := s.loops[s.nextLoopIdx]
	s.nextLoopIdx = (s.nextLoopIdx + 1) % len(s.loops)
	return loop
}

func (s *Service) Start() {
	wg := sync.WaitGroup{}

	length := len(s.loops)
	for i := 0; i < length; i++ {
		wg.Add(1)
		loops := s.loops[i]
		go func() {
			log.Printf("loops run err %s\n", loops.Run())
			wg.Done()
		}()
	}

	wg.Wait()
	log.Printf("service all loops stop \n")
}

// Stop 关闭 Server
func (s *Service) Stop() {
	for k := range s.loops {
		if err := s.loops[k].Stop(); err != nil {
			log.Printf("stop loopEvent err %s \n", err)
		}
	}
}

type EventLoop struct {
	poll    *poller.Poller
	fd2Conn sync.Map //map[int]*AioConn

	mu          sync.Mutex
	pendingFunc []func()
}

func NewEventLoop() (*EventLoop, error) {
	p, err := poller.OpenPoller()
	if err != nil {
		return nil, err
	}
	return &EventLoop{
		fd2Conn: sync.Map{}, //map[int]*AioConn{},
		poll:    p,
	}, nil
}

func (loop *EventLoop) Watch(conn *AioConn) error {
	if _, ok := loop.fd2Conn.Load(conn.fd); !ok {
		_ = syscall.SetNonblock(conn.fd, true)
		loop.fd2Conn.Store(conn.fd, conn)
		return loop.poll.AddRead(conn.fd)
	}
	return errors.New("already in loops")
}

func (loop *EventLoop) Remove(fd int) {
	loop.fd2Conn.Delete(fd)
	_ = loop.poll.Delete(fd)
}

func (loop *EventLoop) Do(fn func()) {
	loop.mu.Lock()
	loop.pendingFunc = append(loop.pendingFunc, fn)
	loop.mu.Unlock()

	if len(loop.pendingFunc) == 1 {
		_ = loop.poll.Trigger()
	}
}

func (loop *EventLoop) Run() error {
	return loop.poll.Polling(func(fd int, ev poller.Event) error {

		// 自己
		if fd == -1 {
			loop.mu.Lock()
			funcs := loop.pendingFunc
			loop.pendingFunc = nil
			loop.mu.Unlock()

			for _, fn := range funcs {
				fn()
			}
			return nil
		}

		if el, ok := loop.fd2Conn.Load(fd); ok {
			if err := el.(*AioConn).handleEvent(fd, ev); err != nil {
				return err
			}
		}
		return nil
	})
}

func (loop *EventLoop) Stop() error {
	loop.fd2Conn.Range(func(key, value interface{}) bool {
		value.(*AioConn).Close("eventLoop stop")
		return true
	})
	return loop.poll.Close()
}
