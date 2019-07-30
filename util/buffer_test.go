package util

import (
	"fmt"
	"testing"
)

func TestNewBuffer(t *testing.T) {
	buffer := NewBuffer(10)
	buffer.Write([]byte{0, 5, 4, 5, 6, 1, 3, 1})
	fmt.Println(buffer.Peek(), buffer.Len())

	u16, err := buffer.ReadUint16BE()
	fmt.Println(u16, err)
	fmt.Println(buffer.Peek(), buffer.Len())

	buffer.WriteUint16BE(56)
	fmt.Println(buffer.Peek(), buffer.Len())

	test, err := buffer.ReadBytes(4)
	fmt.Println(test, err)
	fmt.Println(buffer.Peek(), buffer.Len())

	buffer.ReadUint16BE()
	fmt.Println(test, err)
	fmt.Println(buffer.Peek())

	bt := buffer.Peek()
	bt[0] = 255
	fmt.Println(buffer.Peek())

	c, err := buffer.ReadByte()
	fmt.Println(c, err)
	fmt.Println(buffer.Peek(), buffer.Len())

	buffer.Write([]byte{10, 15, 14, 15, 16, 15, 17, 12, 18})
	fmt.Println(buffer.Peek(), buffer.Len())

}
