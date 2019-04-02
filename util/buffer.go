package util

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

const (
	defBuffSize = 1 << 16 //65536
)

type Buffer struct {
	lock sync.Mutex
	buf  []byte
	woff int //写偏移
}

func NewBuffer() *Buffer {
	return &Buffer{
		buf:  make([]byte, defBuffSize),
		woff: 0,
	}
}

func NewSizeBuffer(length int) *Buffer {
	return &Buffer{
		buf:  make([]byte, length),
		woff: 0,
	}
}

func (b *Buffer) Size() int {
	return b.woff
}

func (b *Buffer) Reset(n int) {
	copy(b.buf, b.buf[n:b.woff])
	b.woff = b.woff - n
}

func (b *Buffer) Reader(r io.Reader) (int, error) {
	n, err := r.Read(b.buf[b.woff:])
	b.woff += n
	return n, err
}

func (b *Buffer) AppendByte(bt []byte) {
	copy(b.buf[b.woff:], bt)
	b.woff += len(bt)
}

func (b *Buffer) AppendUint16(num uint16) {
	binary.BigEndian.PutUint16(b.buf[b.woff:], num)
	b.woff += 2
}

func (b *Buffer) GetUint16() (uint16, error) {
	if b.woff < 2 {
		return 0, fmt.Errorf("data not enough")
	} else {
		return binary.BigEndian.Uint16(b.buf[:2]), nil
	}
}

func (b *Buffer) GetBytes(index, size int) []byte {
	return b.buf[index : index+size]
}
