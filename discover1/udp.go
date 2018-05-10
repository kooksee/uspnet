// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package discover

import (
	"container/list"
	"crypto/ecdsa"
	"net"
	"time"

	"github.com/kooksee/uspnet/log"
	"github.com/kooksee/uspnet/p2p/netutil"
	"github.com/kooksee/uspnet/rlp"
)

// udp implements the RPC protocol.
type udp struct {
	conn        conn
	netRestrict *netutil.Netlist
	priV        *ecdsa.PrivateKey
	ourEndpoint rpcEndpoint

	addPending chan *pending
	gotReply   chan reply

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
type pending struct {
	// these fields must match in the reply.
	from  NodeID
	pType byte

	// time when the request must complete
	deadline time.Time

	// callback is called when a matching reply arrives. if it returns
	// true, the callback is removed from the pending reply queue.
	// if it returns false, the reply is considered incomplete and
	// the callback will be invoked again for the next matching reply.
	callback func(resp interface{}) (done bool)

	// errc receives nil when the callback indicates completion or an
	// error if no further reply is received within the timeout.
	errC chan<- error
}

type reply struct {
	from  NodeID
	pType byte
	data  interface{}
	// loop indicates whether there was
	// a matching request by sending on this channel.
	matched chan<- bool
}

// ListenUDP returns a new table that listens for UDP packets on laddr.
func ListenUDP(priV *ecdsa.PrivateKey, lAddr string, nodeDBPath string, netRestrict *netutil.Netlist) (*Table, error) {
	addr, err := net.ResolveUDPAddr("udp", lAddr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	tab, _, err := newUDP(priV, conn, nodeDBPath, netRestrict)
	if err != nil {
		return nil, err
	}
	log.Info("UDP listener up", "self", tab.self)
	return tab, nil
}

func newUDP(priV *ecdsa.PrivateKey, c conn, nodeDBPath string, netRestrict *netutil.Netlist) (*Table, *udp, error) {
	udp := &udp{
		conn:        c,
		priV:        priV,
		netRestrict: netRestrict,
		closing:     make(chan struct{}),
		gotReply:    make(chan reply),
		addPending:  make(chan *pending),
	}
	realAddr := c.LocalAddr().(*net.UDPAddr)

	udp.ourEndpoint = makeEndpoint(realAddr)
	tab, err := newTable(udp, PubkeyID(&priV.PublicKey), realAddr, nodeDBPath)
	if err != nil {
		return nil, nil, err
	}
	udp.Table = tab

	go udp.loop()
	go udp.readLoop()
	return udp.Table, udp, nil
}

func (t *udp) close() {
	close(t.closing)
	t.conn.Close()
}

// ping sends a ping message to the given node and waits for a reply.
func (t *udp) ping(toId NodeID, toAddr *net.UDPAddr) error {
	p := &ping{
		Version:    Version,
		From:       t.ourEndpoint,
		To:         makeEndpoint(toAddr),
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	}

	errC := t.pending(toId, p.Id(), func(interface{}) bool { return true })
	t.send(toAddr, p)
	return <-errC
}

func (t *udp) waitPing(from NodeID) error {
	return <-t.pending(from, (&ping{}).Id(), func(interface{}) bool { return true })
}

func (t *udp) FindNode(toId NodeID, toAddr *net.UDPAddr, target NodeID) ([]*Node, error) {
	return t.findNode(toId, toAddr, target)
}

// findnode sends a findnode request to the given node and waits until
// the node has sent up to k neighbors.
func (t *udp) findNode(toId NodeID, toAddr *net.UDPAddr, target NodeID) ([]*Node, error) {
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
			n, err := t.nodeFromRPC(toAddr, rn)
			if err != nil {
				log.Trace("Invalid neighbor node received", "ip", rn.IP, "addr", toAddr, "err", err.Error())
				continue
			}
			nodes = append(nodes, n)
		}
		return nReceived >= bucketSize
	})
	t.send(toAddr, p)
	return nodes, <-errC
}

// pending adds a reply callback to the pending reply queue.
// see the documentation of type pending for a detailed explanation.
func (t *udp) pending(id NodeID, pt byte, callback func(interface{}) bool) <-chan error {
	ch := make(chan error, 1)
	p := &pending{from: id, pType: pt, callback: callback, errC: ch}
	select {
	case t.addPending <- p:
		// loop will handle it
	case <-t.closing:
		ch <- errClosed
	}
	return ch
}

func (t *udp) handleReply(from NodeID, pt byte, req Packet) bool {
	matched := make(chan bool, 1)
	select {
	case t.gotReply <- reply{from, pt, req, matched}:
		// loop will handle it
		return <-matched
	case <-t.closing:
		return false
	}
}

// loop runs in its own goroutine. it keeps track of
// the refresh timer and the pending reply queue.
func (t *udp) loop() {
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

	for {
		resetTimeout()

		select {
		case <-t.closing:
			for el := plist.Front(); el != nil; el = el.Next() {
				el.Value.(*pending).errC <- errClosed
			}
			return

		case p := <-t.addPending:
			p.deadline = time.Now().Add(respTimeout)
			plist.PushBack(p)

		case r := <-t.gotReply:
			var matched bool
			for el := plist.Front(); el != nil; el = el.Next() {
				p := el.Value.(*pending)
				if p.from == r.from && p.pType == r.pType {
					matched = true
					// Remove the matcher if its callback indicates
					// that all replies have been received. This is
					// required for packet types that expect multiple
					// reply packets.
					if p.callback(r.data) {
						p.errC <- nil
						plist.Remove(el)
					}
					// Reset the continuous timeout counter (time drift detection)
					contTimeouts = 0
				}
			}
			r.matched <- matched

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

func init() {
	p := neighbors{Expiration: ^uint64(0)}
	maxSizeNode := rpcNode{IP: make(net.IP, 16), UDP: ^uint16(0)}
	for n := 0; ; n++ {
		p.Nodes = append(p.Nodes, maxSizeNode)
		size, _, err := rlp.EncodeToReader(p)
		if err != nil {
			// If this ever happens, it will be caught by the unit tests.
			panic("cannot encode: " + err.Error())
		}
		if headSize+size+1 >= 1280 {
			maxNeighbors = n
			break
		}
	}
}

func (t *udp) send(toAddr *net.UDPAddr, req Packet) error {
	if packet, err := encodePacket(t.priV, req.Id(), req); err != nil {
		return err
	} else {
		_, err = t.conn.WriteToUDP(packet, toAddr)
		log.Trace(">> "+req.String(), "addr", toAddr, "err", err.Error())
		return err
	}

	return nil
}

// readLoop runs in its own goroutine. it handles incoming UDP packets.
func (t *udp) readLoop() {
	defer t.conn.Close()
	// Discovery packets are defined to be no larger than 1280 bytes.
	// Packets larger than this size will be cut at the end and treated
	// as invalid because their hash won't match.
	buf := make([]byte, 1280)
	for {
		nBytes, from, err := t.conn.ReadFromUDP(buf)
		if netutil.IsTemporaryError(err) {
			// Ignore temporary read errors.
			log.Debug("Temporary UDP read error", "err", err)
			continue
		} else if err != nil {
			// Shut down the loop for permament errors.
			log.Debug("UDP read error", "err", err)
			return
		}
		t.handlePacket(from, buf[:nBytes])
	}
}

func (t *udp) handlePacket(from *net.UDPAddr, buf []byte) error {
	packet, fromID, hash, err := decodePacket(buf)
	if err != nil {
		log.Debug("Bad discv4 packet", "addr", from, "err", err)
		return err
	}

	err = packet.OnHandle(t, from, fromID, hash)
	log.Trace("<< "+packet.String(), "addr", from, "err", err)
	return err
}

func (t *udp) broadcast(p broadcast) {
	//t.ReadRandomNodes()
	//t.send()
}
func (t *udp) uniCast(p unicast) {
	//t.send()
}
func (t *udp) multiCast(p multicast) {
	//t.send()
}

//realAddr.IP.IsLoopback()
//if natM != nil {
//	if !realAddr.IP.IsLoopback() {
//		go nat.Map(natM, udp.closing, "udp", realAddr.Port, realAddr.Port, "ethereum discovery")
//	}
//	// TODO: react to external IP changes over time.
//	if ext, err := natM.ExternalIP(); err == nil {
//		realAddr = &net.UDPAddr{IP: ext, Port: realAddr.Port}
//	}
//}
