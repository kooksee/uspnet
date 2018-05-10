package discover

import "net"

type Packet interface {
	OnHandle(t *udp, from *net.UDPAddr, fromID NodeID, mac []byte) error
	Id() byte
	String() string
}

type conn interface {
	ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error)
	WriteToUDP(b []byte, addr *net.UDPAddr) (n int, err error)
	Close() error
	LocalAddr() net.Addr
}

// transport is implemented by the UDP transport.
// it is an interface so we can test without opening lots of UDP
// sockets and without generating a private key.
type transport interface {
	ping(NodeID, *net.UDPAddr) error
	waitPing(NodeID) error
	findNode(NodeID, *net.UDPAddr, NodeID) ([]*Node, error)
	close()
}
