package dht

import (
	"crypto/sha1"
	"math/rand"
	"time"
)

type Secret struct {
	cur  []byte
	old  []byte
	rand *rand.Rand
}

func NewSecret() (s *Secret) {
	s = &Secret{
		cur:  make([]byte, 8),
		old:  make([]byte, 8),
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	s.Update()
	return
}

func (s *Secret) Update() {
	if len(s.cur) != 0 {
		copy(s.old, s.cur)
	}
	s.rand.Read(s.cur)
	if len(s.old) == 0 {
		copy(s.old, s.cur)
	}
}

func (s *Secret) Create(addr string) []byte {
	return s.create(addr, false)
}

func (s *Secret) create(addr string, old bool) []byte {
	sec := s.cur
	if old {
		sec = s.old
	}
	h := sha1.New()
	h.Write(sec)
	h.Write([]byte(addr))
	return h.Sum(nil)
}

func (s *Secret) Match(addr string, token []byte) bool {
	t := string(token)
	t1 := s.create(addr, false)
	if string(t1) == t {
		return true
	}
	t2 := s.create(addr, true)
	if string(t2) == t {
		return true
	}
	return false
}
