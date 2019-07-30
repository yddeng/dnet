package util

import (
	"encoding/binary"
	"errors"
	"io"
	"unsafe"
)

var (
	ErrWriteNone = errors.New("buffer: write none")
	ErrFull      = errors.New("buffer: full")
	ErrData      = errors.New("buffer: data not enough")
)

type Buffer struct {
	buff       []byte
	cap        int
	roff, woff int //读、写偏移
}

//由于消息头的长度uint16的原因,故cap的最大值为65535
func NewBuffer(cap int) *Buffer {
	return &Buffer{
		buff: make([]byte, cap),
		cap:  cap,
		roff: 0,
		woff: 0,
	}
}

//已用
func (b *Buffer) Len() int {
	return b.woff - b.roff
}

//未用
func (b *Buffer) UsableLen() int {
	return b.cap - (b.woff - b.roff)
}

func (b *Buffer) reset() {
	copy(b.buff, b.buff[b.roff:b.woff])
	b.woff = b.woff - b.roff
	b.roff = 0
}

//拷贝一份数据，不重置
func (b *Buffer) Peek() []byte {
	var ret = make([]byte, b.woff-b.roff)
	copy(ret, b.buff[b.roff:b.woff])
	return ret
}

func (b *Buffer) Clear() {
	b.buff = make([]byte, b.cap)
	b.woff = 0
	b.roff = 0
}

func (b *Buffer) ReadFrom(reader io.Reader) (int, error) {
	//重置缓存区
	b.reset()

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

// 写入缓冲区
func (b *Buffer) Write(bytes []byte) (n int, err error) {
	var needSz = len(bytes)
	if 0 == needSz {
		return 0, ErrWriteNone
	}

	// 可写入大小
	var usableLen = b.UsableLen()
	//fmt.Println(usableLen)

	// 缓冲满, 或写不下
	if 0 == usableLen || needSz > usableLen {
		return 0, ErrFull
	} else {
		//连续可写入区域不够，将重置缓存区
		var writeLen = b.cap - b.woff
		if needSz > writeLen {
			b.reset()
		}
	}

	// 计算可连续写入缓冲大小
	var endPos = b.woff + needSz
	//fmt.Println(endPos)

	// 写入缓冲
	copy(b.buff[b.woff:endPos], bytes)

	// 更新写入位置
	b.woff += needSz

	return needSz, nil
}

func (b *Buffer) WriteUint16BE(num uint16) {
	var bt = make([]byte, 2)
	binary.BigEndian.PutUint16(bt, num)
	b.Write(bt)
}

func (b *Buffer) ReadUint16BE() (uint16, error) {
	num, err := b.ReadBytes(2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(num), nil
}

func (b *Buffer) WriteUint32BE(num uint32) {
	var bt = make([]byte, 4)
	binary.BigEndian.PutUint32(bt, num)
	b.Write(bt)
}

func (b *Buffer) ReadUint32BE() (uint32, error) {
	num, err := b.ReadBytes(4)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(num), nil
}

func (b *Buffer) WriteUint64BE(num uint64) {
	var bt = make([]byte, 8)
	binary.BigEndian.PutUint64(bt, num)
	b.Write(bt)
}

func (b *Buffer) ReadUint64BE() (uint64, error) {
	num, err := b.ReadBytes(8)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(num), nil
}

func (b *Buffer) WriteBytes(data []byte) {
	b.Write(data)
}

//获取len长度的数据
func (b *Buffer) ReadBytes(len int) ([]byte, error) {
	end := b.roff + len
	if end > b.woff {
		return nil, ErrData
	}

	var ret = make([]byte, len)
	copy(ret, b.buff[b.roff:end])
	b.roff += len

	return ret, nil
}

func (b *Buffer) ReadByte() (byte, error) {
	if b.Len() < 1 {
		return 0, ErrData
	}

	ret := b.buff[b.roff]
	b.roff++

	return ret, nil
}

func (b *Buffer) WriteByte(c byte) {
	b.Write([]byte{c})
}

func (b *Buffer) WriteString(str string) {
	data := []byte(str)
	b.Write(data)
}

//获取len长度的数据
func (b *Buffer) ReadString(len int) (string, error) {
	bytes, err := b.ReadBytes(len)
	if err != nil {
		return "", err
	}

	//不用拷贝
	ret := *(*string)(unsafe.Pointer(&bytes))

	return ret, nil
}
