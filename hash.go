package dht

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
)

type hash []byte

const hashLen int = 20

func newHash(b []byte) (hash, error) {
	h := make([]byte, hashLen)
	n, err := hex.Decode(h, b)
	if err != nil {
		return nil, err
	}
	if n != hashLen {
		str := fmt.Sprintf("the length of the \"%s\" is not equal to %d.", string(b), hashLen)
		return h, errors.New(str)
	}
	return h, nil
}

func newRandomHash() (hash, error) {
	buf := bytes.NewBuffer(nil)
	for i := 0; i < hashLen; i++ {
		buf.WriteString(fmt.Sprintf("%02x", rand.Intn(256)))
	}
	return newHash(buf.Bytes())
}

func (h hash) compare(o hash) int {
	return bytes.Compare(h, o)
}

func (h hash) distance(o hash) hash {
	var m [hashLen]byte
	for i := 0; i < hashLen; i++ {
		m[i] = h[i] ^ o[i]
	}
	return hash(m[:])
}

func (h hash) middle(o hash) hash {
	var rem int
	var m [hashLen]byte
	for i := 0; i < hashLen; i++ {
		n := rem + int(h[i]) + int(o[i])
		if rem = n % 2; rem == 1 {
			rem = 256
		}
		m[i] = byte(n / 2)
	}
	return hash(m[:])
}

func (h hash) inRange(min, max hash) bool {
	return h.compare(min) >= 0 && h.compare(max) < 0
}

func (h hash) String() string {
	return fmt.Sprintf("%02x", string(h))
}
