package dnet

import "net"

//编码器
type Encode interface {
	Pack(interface{}) ([]byte, error)
}

type Reader interface {
	Receive(net.Conn) (interface{}, error)
}
