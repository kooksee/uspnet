package discover

import "net"

func makeEndpoint(addr *net.UDPAddr) rpcEndpoint {
	ip := addr.IP.To4()
	if ip == nil {
		ip = addr.IP.To16()
	}
	return rpcEndpoint{IP: ip, UDP: uint16(addr.Port)}
}

