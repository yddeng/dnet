package tcp

import "net"

func NewTCPListener(addr string) (*net.TCPListener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	return net.ListenTCP("tcp", tcpAddr)
}

func NewTcpConnector(addr string) (net.Conn, error) {
	return net.Dial("tcp", addr)
}
