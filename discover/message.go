package discover

import (
	"crypto/ecdsa"
	"bytes"
	"fmt"
	"github.com/kooksee/uspnet/rlp"
	"github.com/kooksee/uspnet/log"
	"github.com/kooksee/uspnet/crypto"
)

const (
	macSize  = 256 / 8
	sigSize  = 520 / 8
	headSize = macSize + sigSize // space of packet frame data
)

var (
	headSpace = make([]byte, headSize)

	// Neighbors replies are sent across multiple packets to
	// stay below the 1280 byte limit. We compute the maximum number
	// of entries by stuffing a packet until it grows too large.
	maxNeighbors int
)

func encodePacket(priv *ecdsa.PrivateKey, pt byte, req interface{}) ([]byte, error) {
	b := new(bytes.Buffer)
	b.Write(headSpace)
	b.WriteByte(pt)
	if err := rlp.Encode(b, req); err != nil {
		log.Error("Can't encode discv4 packet", "err", err.Error())
		return nil, err
	}

	packet := b.Bytes()
	sig, err := crypto.Sign(crypto.Keccak256(packet[headSize:]), priv)
	if err != nil {
		log.Error("Can't sign discv4 packet", "err", err)
		return nil, err
	}

	copy(packet[macSize:], sig)
	copy(packet, crypto.Keccak256(packet[macSize:]))
	return packet, nil
}

func decodePacket(buf []byte) (Packet, NodeID, []byte, error) {
	if len(buf) < headSize+1 {
		return nil, NodeID{}, nil, errPacketTooSmall
	}

	hash, sig, sigData := buf[:macSize], buf[macSize:headSize], buf[headSize:]
	if !bytes.Equal(hash, crypto.Keccak256(buf[macSize:])) {
		return nil, NodeID{}, nil, errBadHash
	}

	fromID, err := recoverNodeID(crypto.Keccak256(buf[headSize:]), sig)
	if err != nil {
		return nil, NodeID{}, hash, err
	}

	req := PacketManager.Packet(sigData[0])
	if req == nil {
		return nil, fromID, hash, fmt.Errorf("unknown type: %d", sigData[0])
	}

	return req, fromID, hash, rlp.NewStream(bytes.NewReader(sigData[1:]), 0).Decode(req)
}
