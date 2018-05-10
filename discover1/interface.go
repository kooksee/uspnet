package discover

import (
	"net"
	"github.com/vmihailenco/msgpack"
	"github.com/kooksee/uspnet/crypto"
	"crypto/ecdsa"
	"time"
)

type IPacket interface {
	OnHandle(t *KRpc, tx *Tx) error
	Type() byte
	String() string
	Encode() ([]byte, error)
	Decode([]byte) error
	Expiration() int64
	Sign(*ecdsa.PrivateKey) ([]byte, error)
	Verify([]byte) error
}

type Packet struct {
	_msgpack struct{} `msgpack:",omitempty"`
	NID      NodeID
	Payload  []byte
	From     rpcEndpoint
	N        uint8 //衰减次数,每传播一次,次数减少1
}

func (p *Packet) OnHandle(t *KRpc, tx *Tx) error { return nil }
func (p *Packet) Type() byte                     { return 0x0 }
func (p *Packet) String() string                 { return "" }
func (p *Packet) Expiration() int64              { return int64(15 * time.Second) }
func (p *Packet) Encode() ([]byte, error) {
	return msgpack.Marshal(p)
}
func (p *Packet) Decode(data []byte) error {
	return msgpack.Unmarshal(data, p)
}
func (p *Packet) Sign(priV *ecdsa.PrivateKey) ([]byte, error) {
	d, err := p.Encode()
	if err != nil {
		return nil, err
	}

	sig, err := crypto.Sign(crypto.Keccak256(d), priV)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func (p *Packet) Verify(sig []byte) error {
	d, err := p.Encode()
	if err != nil {
		return err
	}

	nid, err := RecoverNodeID(crypto.Keccak256(d), sig)
	if err != nil {
		return err
	}
	p.NID = nid
	return nil
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
	Ping(NodeID, *net.UDPAddr) error
	WaitPing(NodeID) error
	FindNode(NodeID, *net.UDPAddr, NodeID) ([]*Node, error)
	Close()
}
