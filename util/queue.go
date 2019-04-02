package util

import (
	"fmt"
	"sync"
)

/*
 * 消息队列，支持动态扩展内存
 * 设置内存最大值，超出后抛出异常，说明外部逻辑处理消息能力低于消息添加
 * 用读写偏移以减少大量的内存拷贝
 */

const minQueueLen = 1 << 4 //16
const maxQueueLen = 1 << 8 //256

type Queue struct {
	buf        []interface{}
	size, cap  int
	roff, woff int //读写偏移
	lock       sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		buf: make([]interface{}, minQueueLen),
		cap: minQueueLen,
	}
}

func (q *Queue) Size() int {
	defer q.lock.Unlock()
	q.lock.Lock()

	return q.size
}

//cap不够，重新分配大小为原来的2倍
func (q *Queue) recap() (err error) {
	if q.cap == maxQueueLen {
		return fmt.Errorf("maxQueueLen")
	}

	newBuf := make([]interface{}, q.cap<<1)
	copy(newBuf, q.buf[:q.size])
	q.buf = newBuf
	q.cap = cap(newBuf)
	q.roff = 0
	q.woff = q.size

	return nil
}

//执行拷贝，将读写偏移移至头部
func (q *Queue) reset() {
	copy(q.buf[:q.size], q.buf[q.roff:q.woff])
	q.roff = 0
	q.woff = q.size
}

func (q *Queue) Add(elem interface{}) (err error) {
	defer q.lock.Unlock()
	q.lock.Lock()

	if q.woff == q.cap {
		if q.size < q.cap {
			q.reset()
		} else {
			err = q.recap()
			if err != nil {
				return
			}
		}
	}

	q.buf[q.woff] = elem
	q.woff++
	q.size++

	return
}

//获取队首
func (q *Queue) GetFront() (elem interface{}, err error) {
	defer q.lock.Unlock()
	q.lock.Lock()

	if q.size == 0 || q.roff >= q.woff {
		return nil, fmt.Errorf("none")
	} else {
		elem = q.buf[q.roff]
		q.roff++
		q.size--
		return
	}
}

func (q *Queue) GetAll() (elems []interface{}) {
	defer q.lock.Unlock()
	q.lock.Lock()

	elems = q.buf[q.roff:q.woff]
	q.clear()
	return
}

func (q *Queue) clear() {
	q.buf = make([]interface{}, minQueueLen)
	q.size = 0
	q.cap = cap(q.buf)
	q.roff = 0
	q.woff = 0
}

func (q *Queue) P() {
	fmt.Println(q.buf, q.size, q.cap, q.roff, q.woff)
}
