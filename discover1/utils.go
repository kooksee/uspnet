package discover

import (
	"net"
	"fmt"
	"github.com/kooksee/uspnet/crypto/secp256k1"
	"github.com/kooksee/uspnet/common"
	"github.com/satori/go.uuid"
	"time"
	"bytes"
	"sync"
)

func MakeEndpoint(addr *net.UDPAddr) rpcEndpoint {
	ip := addr.IP.To4()
	if ip == nil {
		ip = addr.IP.To16()
	}
	return rpcEndpoint{IP: ip, UDP: uint16(addr.Port)}
}

// recoverNodeID computes the public key used to sign the
// given hash from the signature.
func RecoverNodeID(hash, sig []byte) (id NodeID, err error) {
	pubKey, err := secp256k1.RecoverPubkey(hash, sig)
	if err != nil {
		return id, err
	}
	if len(pubKey)-1 != len(id) {
		return id, fmt.Errorf("recovered pubkey has %d bits, want %d bits", len(pubKey)*8, (len(id)+1)*8)
	}
	for i := range id {
		id[i] = pubKey[i+1]
	}
	return id, nil
}

// DistCmp compares the distances a->target and b->target.
// Returns -1 if a is closer to target, 1 if b is closer to target
// and 0 if they are equal.
func DistCmp(target, a, b common.Hash) int {
	for i := range target {
		da := a[i] ^ target[i]
		db := b[i] ^ target[i]
		if da > db {
			return 1
		} else if da < db {
			return -1
		}
	}
	return 0
}

func UUID() []byte {
	for {
		uid, err := uuid.NewV4()
		if err == nil {
			return uid.Bytes()
		}
	}
	return nil
}

func expired(ts int64) bool {
	return time.Unix(ts, 0).Before(time.Now())
}

func timeAdd(ts time.Duration) time.Time {
	return time.Now().Add(ts)
}

type KBuffer struct {
	buf   map[string][]byte
	Delim []byte

	sync.RWMutex
}

func (t *KBuffer) Add(key string, b []byte) {
	t.Lock()
	defer t.Unlock()

	t.buf[key] = append(t.buf[key], b...)
}

func (t *KBuffer) Next(key string) [][]byte {
	t.RLock()
	defer t.RUnlock()

	if len(t.buf) != 0 {
		d := bytes.Split(t.buf[key], t.delim)
		if len(d) > 1 {
			t.buf[key] = d[len(d)-1]
			return d[:len(d)-2]
		}
	}
	return nil
}
