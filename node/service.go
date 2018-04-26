package node

import "github.com/kooksee/uspnet/utils/net"

type Service struct {
	conn *net.Conn
}

type Services struct {
	sList []*Service
}
