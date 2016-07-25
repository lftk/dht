package dht

import (
	"crypto/sha1"
	"math/rand"
	"time"
)

type secret struct {
	cur  []byte
	old  []byte
	rand *rand.Rand
}

func newSecret() (s *secret) {
	s = &secret{
		cur:  make([]byte, 8),
		old:  make([]byte, 8),
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	s.rand.Read(s.cur)
	copy(s.old, s.cur)
	return
}

func (s *secret) Update() {
	copy(s.old, s.cur)
	s.rand.Read(s.cur)
}

func (s *secret) Create(b []byte) []byte {
	return s.create(b, false)
}

func (s *secret) create(b []byte, old bool) []byte {
	sec := s.cur
	if old {
		sec = s.old
	}
	h := sha1.New()
	h.Write(sec)
	h.Write(b)
	return h.Sum(nil)
}

func (s *secret) Match(b, token []byte) bool {
	t := string(token)
	t1 := s.create(b, false)
	if string(t1) == t {
		return true
	}
	t2 := s.create(b, true)
	if string(t2) == t {
		return true
	}
	return false
}
