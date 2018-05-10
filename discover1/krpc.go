package discover

import (
	"container/list"
	"crypto/ecdsa"
	"net"
	"time"
	"io"

	"github.com/kooksee/uspnet/log"
	"github.com/kooksee/uspnet/p2p/netutil"

	"github.com/vmihailenco/msgpack"
)

// udp implements the RPC protocol.
type KRpc struct {
	conn        conn
	netRestrict *netutil.Netlist
	priV        *ecdsa.PrivateKey
	ourEndpoint rpcEndpoint

	sendChan chan *sendQ
	recvChan chan *recvQ

	closing chan struct{}
	*Table
}

// pending represents a pending reply.
//
// some implementations of the protocol wish to send more than one
// reply packet to findnode. in general, any neighbors packet cannot
// be matched up with a specific findnode packet.
//
// our implementation handles this by storing a callback function for
// each pending reply. incoming packets from a node are dispatched
// to all the callback functions for that node.

type sendQ struct {
	tx       *Tx
	deadline time.Time
}

type recvQ struct {
	tx       *Tx
	deadline time.Time
	buf      []byte
}

// ListenUDP returns a new table that listens for UDP packets on laddr.
func NewKRpc(priV *ecdsa.PrivateKey, lAddr string, nodeDBPath string, netRestrict *netutil.Netlist) (*KRpc, error) {
	addr, err := net.ResolveUDPAddr("udp", lAddr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	krpc, err := newKRpc(priV, conn, nodeDBPath, netRestrict)
	if err != nil {
		return nil, err
	}
	log.Info("UDP listener up", "self", krpc.self)
	return krpc, nil
}

func newKRpc(priV *ecdsa.PrivateKey, c conn, nodeDBPath string, netRestrict *netutil.Netlist) (*KRpc, error) {
	kRpc := &KRpc{
		conn:        c,
		priV:        priV,
		netRestrict: netRestrict,
		closing:     make(chan struct{}),
		sendChan:    make(chan *sendQ, 10000),
		recvChan:    make(chan *recvQ, 10000),
	}

	// 发送缓存和接受缓存
	//kRpc.sendCache = cache.New(5*time.Minute, 10*time.Minute)
	//kRpc.recvCache = cache.New(5*time.Minute, 10*time.Minute)

	realAddr := c.LocalAddr().(*net.UDPAddr)

	kRpc.ourEndpoint = MakeEndpoint(realAddr)
	tab, err := newTable(kRpc, PubkeyID(&priV.PublicKey), realAddr, nodeDBPath)
	if err != nil {
		return nil, err
	}
	kRpc.Table = tab

	go kRpc.loop()
	go kRpc.readLoop()
	go kRpc.writeLoop()
	return kRpc, nil
}

func (t *KRpc) close() {
	close(t.closing)
	t.conn.Close()
}

// ping sends a ping message to the given node and waits for a reply.
func (t *KRpc) Ping(toId NodeID, toAddr *net.UDPAddr) error {
	p := &ping{Version: Version, To: MakeEndpoint(toAddr)}
	p.From = t.ourEndpoint

	return t.Unicast(&Tx{
		Type: p.Type(),
	})
}

// findnode sends a findnode request to the given node and waits until
// the node has sent up to k neighbors.
func (t *KRpc) FindNode(toId NodeID, toAddr *net.UDPAddr, target NodeID) ([]*Node, error) {
	nodes := make([]*Node, 0, bucketSize)
	nReceived := 0

	p := &findNode{
		Target:     target,
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	}

	errC := t.pending(toId, p.Id(), func(r interface{}) bool {
		reply := r.(*neighbors)
		for _, rn := range reply.Nodes {
			nReceived++
			n, err := t.NodeFromRPC(toAddr, rn)
			if err != nil {
				log.Trace("Invalid neighbor node received", "ip", rn.IP, "addr", toAddr, "err", err.Error())
				continue
			}
			nodes = append(nodes, n)
		}
		return nReceived >= bucketSize
	})
	t.Send(toAddr, p)
	return nodes, <-errC
}

// loop runs in its own goroutine. it keeps track of
// the refresh timer and the pending reply queue.
func (t *KRpc) loop() {
	var (
		plist        = list.New()
		timeout      = time.NewTimer(0)
		nextTimeout  *pending // head of plist when timeout was last reset
		contTimeouts = 0      // number of continuous timeouts to do NTP checks
		ntpWarnTime  = time.Unix(0, 0)
	)
	<-timeout.C // ignore first timeout
	defer timeout.Stop()

	resetTimeout := func() {
		if plist.Front() == nil || nextTimeout == plist.Front().Value {
			return
		}
		// Start the timer so it fires when the next pending reply has expired.
		now := time.Now()
		for el := plist.Front(); el != nil; el = el.Next() {
			nextTimeout = el.Value.(*pending)
			if dist := nextTimeout.deadline.Sub(now); dist < 2*respTimeout {
				timeout.Reset(dist)
				return
			}
			// Remove pending replies whose deadline is too far in the
			// future. These can occur if the system clock jumped
			// backwards after the deadline was assigned.
			nextTimeout.errC <- errClockWarp
			plist.Remove(el)
		}
		nextTimeout = nil
		timeout.Stop()
	}

	handleTx := func(buf []byte, tx *Tx) error {
		if err := msgpack.Unmarshal(buf, tx); err != nil {
			return err
		}
		if err := PacketManager.Packet(tx.Type).Verify(tx.SignMsg); err != nil {
			return err
		}
		return PacketManager.Packet(tx.Type).OnHandle(t, tx)
	}

	for {
		resetTimeout()

		select {
		case p := <-t.recvChan:
			if expired(p.deadline.Unix()) {
				continue
			}

			if err := handleTx(p.buf, p.tx); err != nil {
				log.Error(err.Error())
				t.recvChan <- p
			}

		case now := <-timeout.C:
			nextTimeout = nil

			// Notify and remove callbacks whose deadline is in the past.
			for el := plist.Front(); el != nil; el = el.Next() {
				p := el.Value.(*pending)
				if now.After(p.deadline) || now.Equal(p.deadline) {
					p.errC <- errTimeout
					plist.Remove(el)
					contTimeouts++
				}
			}
			// If we've accumulated too many timeouts, do an NTP time sync check
			if contTimeouts > ntpFailureThreshold {
				if time.Since(ntpWarnTime) >= ntpWarningCooldown {
					ntpWarnTime = time.Now()
					go checkClockDrift()
				}
				contTimeouts = 0
			}
		}
	}
}

func (t *KRpc) send(tx *Tx) error {
	p := PacketManager.Packet(tx.Type)
	sig, err := p.Sign(t.priV)
	if err != nil {
		return err
	}

	tx.SignMsg = sig
	d, err := tx.Encode()
	if err != nil {
		return err
	}

	tx.data = d
	t.sendChan <- &sendQ{
		tx:       tx,
		deadline: timeAdd(sendTimeout),
	}

	return nil
}

func (t *KRpc) writeLoop() {
	for {
		p := <-t.sendChan
		if expired(p.deadline.Unix()) {
			// 发送过期提醒,过期跳过不执行了
			continue
		}

		switch p.tx.netType {
		case UNICAST:
			uc := PacketManager.Packet(p.tx.Type).(Unicast)
			t.writeUnicast(p.tx.data, uc.To.addr())
		case MULTICAST:
			uc := PacketManager.Packet(p.tx.Type).(Multicast)
			t.writeMulticast(p.tx.data, uc.To...)
		case BROADCAST:
			t.writeBroadcast(p.tx.data)
		}
	}
}

func (t *KRpc) writeUnicast(d []byte, toAddr *net.UDPAddr) {
	_, err := t.conn.WriteToUDP(d, toAddr)
	if err == nil {
		return
	}
	log.Error("sent unicast error", "toAddr", toAddr.String())

	if t.LenTab() == 0 {
		log.Error("节点列表为空")
		return
	}

	nodes := make([]*Node, 8)
	if t.ReadRandomNodes(nodes) == 0 {
		log.Error("找不到节点")
		return
	}

	for _, node := range nodes {
		_, err = t.conn.WriteToUDP(d, node.addr())
		log.Error("sent unicast error", "toAddr", node.addr().String())
	}
}

func (t *KRpc) writeMulticast(d []byte, toAddrs ... rpcEndpoint) {
	for _, addr := range toAddrs {
		t.writeUnicast(d, addr.addr())
	}
}

func (t *KRpc) writeBroadcast(d []byte) {
	if t.LenTab() == 0 {
		log.Error("节点列表为空")
		return
	}

	nodes := make([]*Node, 16)
	if t.ReadRandomNodes(nodes) == 0 {
		log.Error("找不到节点")
		return
	}

	for _, node := range nodes {
		_, err := t.conn.WriteToUDP(d, node.addr())
		if err != nil {
			log.Error("sent unicast error", "toAddr", node.String())
		}
	}

	return
}

func (t *KRpc) Unicast(tx *Tx) error {
	tx.netType = UNICAST
	return t.send(tx)
}

func (t *KRpc) Multicast(tx *Tx) error {
	tx.netType = MULTICAST
	return t.send(tx)
}

func (t *KRpc) Broadcast(tx *Tx) error {
	tx.netType = BROADCAST
	return t.send(tx)
}

// readLoop runs in its own goroutine. it handles incoming UDP packets.
func (t *KRpc) readLoop() {
	defer t.conn.Close()
	// Discovery packets are defined to be no larger than 1280 bytes.
	// Packets larger than this size will be cut at the end and treated
	// as invalid because their hash won't match.
	kb := KBuffer{Delim: []byte(DELIMITER)}
	for {

		buf := make([]byte, MAX_BUF_LEN)
		nBytes, from, err := t.conn.ReadFromUDP(buf)

		switch err {
		case nil:
			key := from.String()
			kb.Add(key, buf[:nBytes])
			t.handlePacket(from, kb.Next(key))
		case io.EOF:
			log.Error("Rx io.EOF: ", err, ", node id is ", from.String())
		default:
			log.Error("Read connection error ", "err", err, "addr", from.String())
		}
	}
}

func (t *KRpc) handlePacket(from *net.UDPAddr, bufList [][]byte) {
	if bufList == nil {
		return
	}

	for _, buf := range bufList {

		t.recvChan <- &recvQ{
			deadline: timeAdd(respTimeout),
			buf:      buf,
			tx:       &Tx{fromAddr: from},
		}
	}
}
