package dht

import (
	"math/rand"
	"testing"
)

func Test_secret(t *testing.T) {
	s := newSecret()
	addr := make([]byte, 20)
	for i := 0; i < 1000; i++ {
		rand.Read(addr)
		b := matchToken(s, addr)
		if b == false {
			t.Fatal(addr)
		}
	}
}

func matchToken(s *secret, addr []byte) bool {
	return s.Match(addr, s.Create(addr))
}
