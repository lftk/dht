package dht

import (
	"math/rand"
	"testing"
)

func Test_Secret(t *testing.T) {
	s := NewSecret()
	addr := make([]byte, 20)
	for i := 0; i < 1000; i++ {
		rand.Read(addr)
		b := matchToken(s, string(addr))
		if b == false {
			t.Fatal(addr)
		}
	}
}

func matchToken(s *Secret, addr string) bool {
	return s.Match(addr, s.Create(addr))
}
