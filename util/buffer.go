package util

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrWriteNone = errors.New("buffer: write none")
	ErrFull      = errors.New("buffer: full")
	ErrData      = errors.New("buffer: data not enough")
)

type Buffer struct {
	buff []byte
	cap  int
	woff int //写偏移
}

func NewBuffer(cap int) *Buffer {
	return &Buffer{
		buff: make([]byte, cap),
		cap:  cap,
		woff: 0,
	}
}

func (b *Buffer) Len() int {
	return b.woff
}

func (b *Buffer) Reset(n int) {
	if b.woff >= n {
		copy(b.buff, b.buff[n:b.woff])
		b.woff = b.woff - n
	}
}

func (b *Buffer) Clear() {
	b.buff = make([]byte, b.cap)
	b.woff = 0
}

func (b *Buffer) ReadFrom(reader io.Reader) (int, error) {
	// 可写入大小
	var space = b.cap - b.woff

	// 缓冲满
	if 0 == space {
		return 0, ErrFull
	}

	// 计算可连续写入缓冲大小
	var endPos = b.woff + space

	// 写入缓冲
	readLen, err := reader.Read(b.buff[b.woff:endPos])

	// 读取错误
	if err != nil {
		return 0, err
	}

	// 没有读取到东西
	if 0 == readLen {
		return 0, ErrWriteNone
	}

	// 更新写入位置
	b.woff += readLen

	return readLen, nil
}

// 写入缓冲区 -- 注: 缓冲满后, 需要上层自己处理
func (b *Buffer) Write(bytes []byte) (n int, err error) {
	var needSz = len(bytes)
	if 0 == needSz {
		return 0, nil
	}

	// 可写入大小
	var space = b.cap - b.woff

	// 缓冲满, 或写不下
	if 0 == space || needSz > space {
		return 0, ErrFull
	}

	// 计算可连续写入缓冲大小
	var endPos = b.woff + space

	// 写入缓冲
	copy(b.buff[b.woff:endPos], bytes)

	// 更新写入位置
	b.woff += needSz

	return needSz, nil
}

//读uint32
func (b *Buffer) Uint32(bytes []byte) (uint32, error) {
	if len(bytes) < 4 {
		return 0, ErrData
	}
	return binary.BigEndian.Uint32(bytes), nil
}

func (b *Buffer) AppendUint16(num uint16) {
	var bt = make([]byte, 2)
	binary.BigEndian.PutUint16(bt, num)
	b.Write(bt)
}

func (b *Buffer) AppendUint32(num uint32) {
	var bt = make([]byte, 4)
	binary.BigEndian.PutUint32(bt, num)
	b.Write(bt)
}

func (b *Buffer) AppendBytes(data []byte) {
	b.Write(data)
}

//获取从start开始的len长度的数据
func (b *Buffer) ReadBytes(start, len int) []byte {
	end := start + len
	if end > b.woff {
		return nil
	}

	return b.buff[start:end]
}

func (b *Buffer) Bytes() []byte {
	return b.buff[:b.woff]
}
