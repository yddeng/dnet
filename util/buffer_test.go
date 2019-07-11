package util

import (
	"fmt"
	"testing"
)

func TestNewBuffer(t *testing.T) {
	buffer := NewBuffer(10)
	buffer.Write([]byte{0, 0, 0, 5, 4, 5, 6, 1, 3, 1})
	fmt.Println(buffer.Bytes())

	u32, err := buffer.Uint32(buffer.ReadBytes(0, 4))
	fmt.Println(u32, err)

	buffer.Reset(4)
	fmt.Println(buffer.Bytes())

	buffer.AppendUint32(25)
	fmt.Println(buffer.Bytes())

	buffer.AppendUint16(56)
	fmt.Println(buffer.Bytes())

}
