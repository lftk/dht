package dht

import (
	"encoding/hex"
	"errors"
	"fmt"
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

func (h hash) String() string {
	return fmt.Sprintf("%02x", string(h))
}

func (h hash) compare(o hash) int {
	for i := 0; i < hashLen; i++ {
		if h[i] == o[i] {
			continue
		}
		if h[i] < o[i] {
			return -1
		}
		return 1
	}
	return 0
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
