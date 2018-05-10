package discover

import (
	"github.com/vmihailenco/msgpack"
	"net"
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

type Tx struct {
	_msgpack struct{} `msgpack:",omitempty"`

	Type    byte
	SignMsg []byte
	ID      []byte

	netType  NetType
	fromAddr *net.UDPAddr
	fromID   NodeID
	data     []byte
}

func (tx *Tx) Encode() ([]byte, error) {
	pk, err := msgpack.Marshal(tx)
	if err != nil {
		return nil, err
	}

	return pk, nil
	//return append(pk, []byte(Delimiter)...), nil
}

func DecodeTx(data []byte) (*Tx, error) {
	tx := &Tx{}
	err := msgpack.Unmarshal(data, tx)
	if err != nil {
		return nil, err
	}

	tx.Event.Verify()

	return tx, nil
}

func (tx *Tx) String() string {
	return tx.Event.String()
}
