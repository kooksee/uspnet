package discover

import (
	"net"
	"github.com/kooksee/uspnet/rlp"
	"time"
	"github.com/kooksee/uspnet/crypto"
	"github.com/kooksee/uspnet/p2p/netutil"
	"github.com/kataras/iris/core/errors"
)

var PacketManager *Packets

func init() {
	PacketManager = &Packets{}
	PacketManager.Add(&ping{})
	PacketManager.Add(&pong{})
	PacketManager.Add(&findNode{})
	PacketManager.Add(&neighbors{})
}

type Packets struct {
	pmap map[byte]IPacket
}

func (t *Packets) Add(p IPacket) error {
	if _, ok := t.pmap[p.Type()]; ok {
		return errors.New("已经存在")
	}

	t.pmap[p.Type()] = p
	return nil
}

func (t *Packets) Packet(id byte) IPacket {
	if p, ok := t.pmap[id]; ok {
		return p
	}
	return nil
}

// RPC request structures
type (
	// 服务查找请求
	queryReq struct {
		SId        []byte
		SName      []byte
		SNode      rpcEndpoint
		Expiration uint64
	}

	// 服务定义
	service struct {
		name string
		id   string
		desc string
		tags map[string]interface{}
		nid  string
	}

	// 服务查询结果
	queryResp struct {
		Id       []byte
		Services map[string]service
	}

	// 多播用于服务之间的通信
	// 中继用于信息的传递,把消息传递给指定的节点,最终目标节点收到消息
	// 当本网络能够访问目标网络的时候，那么直接发送给目标网络，当本网络无法直接访问的时候，或者访问比较慢的时候
	// 那么就把信息发送给其他的节点，然后中继传播
	Multicast struct {
		*Packet
		To []rpcEndpoint
	}

	// 广播针对所有节点的通知,把消息传递给每一个节点,最终所有节点信息一致
	Broadcast struct {
		*Packet
	}

	// 单播用于一个节点的通信
	Unicast struct {
		*Packet
		To rpcEndpoint
	}

	ping struct {
		*Packet
		Version uint
		To      rpcEndpoint
	}

	// pong is the reply to ping.
	pong struct {
		// This field should mirror the UDP envelope address
		// of the ping packet, which provides a way to discover the
		// the external address (after NAT).
		To rpcEndpoint

		ReplyTok   []byte // This contains the hash of the ping packet.
		Expiration uint64 // Absolute timestamp at which the packet becomes invalid.
		// Ignore additional fields (for forward compatibility).
		Rest []rlp.RawValue `rlp:"tail"`
	}

	// findnode is a query for nodes close to the given target.
	findNode struct {
		Target     NodeID // doesn't need to be an actual public key
		Expiration uint64
		// Ignore additional fields (for forward compatibility).
		Rest []rlp.RawValue `rlp:"tail"`
	}

	// reply to findnode
	neighbors struct {
		Nodes      []rpcNode
		Expiration uint64
		// Ignore additional fields (for forward compatibility).
		Rest []rlp.RawValue `rlp:"tail"`
	}

	rpcNode struct {
		IP  net.IP // len 4 for IPv4 or 16 for IPv6
		UDP uint16 // for discovery protocol
		ID  NodeID
	}

	rpcEndpoint struct {
		IP  net.IP // len 4 for IPv4 or 16 for IPv6
		UDP uint16 // for discovery protocol
	}
)

func (n rpcEndpoint) addr() *net.UDPAddr {
	return &net.UDPAddr{IP: n.IP, Port: int(n.UDP)}
}

func (req *findNode) Id() byte  { return 0x1 }
func (req *neighbors) Id() byte { return 0x2 }
func (req *ping) Id() byte      { return 0x3 }
func (req *pong) Id() byte      { return 0x4 }

func (req *findNode) String() string  { return "findnode" }
func (req *neighbors) String() string { return "neighbors" }
func (req *ping) String() string      { return "ping" }
func (req *pong) String() string      { return "pong" }

func (req *neighbors) OnHandle(t *udp, from *net.UDPAddr, fromID NodeID, mac []byte) error {
	if expired(req.Expiration) {
		return errExpired
	}
	if !t.handleReply(fromID, req.Id(), req) {
		return errUnsolicitedReply
	}
	return nil
}

func (req *ping) OnHandle(t *KRpc, tx *Tx) error {
	if expired(req.Expiration()) {
		return errExpired
	}

	//t.Multicast()

	t.send(from, &pong{
		To:         makeEndpoint(from),
		ReplyTok:   mac,
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	})
	if !t.handleReply(fromID, req.Id(), req) {
		// Note: we're ignoring the provided IP address right now
		go t.bond(true, fromID, from)
	}
	return nil
}

func (req *pong) OnHandle(t *udp, from *net.UDPAddr, fromID NodeID, mac []byte) error {
	if expired(req.Expiration) {
		return errExpired
	}
	if !t.handleReply(fromID, req.Id(), req) {
		return errUnsolicitedReply
	}
	return nil
}

func (req *findNode) OnHandle(t *udp, from *net.UDPAddr, fromID NodeID, mac []byte) error {
	if expired(req.Expiration) {
		return errExpired
	}
	if t.db.node(fromID) == nil {
		// No bond exists, we don't process the packet. This prevents
		// an attack vector where the discovery protocol could be used
		// to amplify traffic in a DDOS attack. A malicious actor
		// would send a findnode request with the IP address and UDP
		// port of the target as the source address. The recipient of
		// the findnode packet would then send a neighbors packet
		// (which is a much bigger packet than findnode) to the victim.
		return errUnknownNode
	}
	target := crypto.Keccak256Hash(req.Target[:])
	t.mutex.Lock()
	closest := t.closest(target, bucketSize).entries
	t.mutex.Unlock()

	p := neighbors{Expiration: uint64(time.Now().Add(expiration).Unix())}
	// Send neighbors in chunks with at most maxNeighbors per packet
	// to stay below the 1280 byte limit.
	for i, n := range closest {
		if netutil.CheckRelayIP(from.IP, n.IP) != nil {
			continue
		}
		p.Nodes = append(p.Nodes, nodeToRPC(n))
		if len(p.Nodes) == maxNeighbors || i == len(closest)-1 {
			t.send(from, &p)
			p.Nodes = p.Nodes[:0]
		}
	}
	return nil
}
