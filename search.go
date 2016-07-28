package dht

import (
	"math"
	"net"
	"time"
)

// CallBack function
type CallBack func(tor *ID, peer []byte)

type node struct {
	id    *ID
	addr  *net.UDPAddr
	time  time.Time
	acked bool
}

type search struct {
	tor   *ID
	cb    CallBack
	nodes map[ID]*node
}

func newSearch(tor *ID, cb CallBack) *search {
	return &search{
		tor:   tor,
		cb:    cb,
		nodes: make(map[ID]*node),
	}
}

func (s *search) Count() int {
	return len(s.nodes)
}

func (s *search) Get(id *ID) *node {
	if n, ok := s.nodes[*id]; ok {
		return n
	}
	return nil
}

func (s *search) Insert(id *ID, addr *net.UDPAddr) (n *node) {
	n, ok := s.nodes[*id]
	if !ok {
		n = &node{
			id:   id,
			addr: addr,
			time: time.Now(),
		}
		s.nodes[*id] = n
	}
	return
}

func (s *search) Remove(id *ID) {
	delete(s.nodes, *id)
}

func (s *search) Notify(tor *ID, peer []byte) {
	if s.cb != nil {
		s.cb(tor, peer)
	}
}

func (s *search) Done(d time.Duration) (done bool) {
	s.Map(func(n *node) bool {
		if n.acked || (d != 0 && time.Since(n.time) > d) {
			done = true
		}
		return done
	})
	return
}

func (s *search) Map(f func(*node) bool) {
	for _, n := range s.nodes {
		if f(n) == false {
			return
		}
	}
}

type searches struct {
	tid int16
	ss  map[int16]*search
}

func newSearches() *searches {
	return &searches{
		ss: make(map[int16]*search),
	}
}

func (s *searches) Count() int {
	return len(s.ss)
}

func (s *searches) Get(tid int16) *search {
	if sr, ok := s.ss[tid]; ok {
		return sr
	}
	return nil
}

func (s *searches) Find(tor *ID) (tid int16, sr *search) {
	tid = -1
	s.Map(func(t int16, s *search) bool {
		if s.tor.Compare(tor) == 0 {
			tid = t
			sr = s
			return false
		}
		return false
	})
	return
}

func (s *searches) nextTID() int16 {
	if n := s.Count(); n < math.MaxInt16+1 {
		for i := 0; i <= n; i++ {
			tid := s.tid
			if s.tid == math.MaxInt16 {
				s.tid = 0
			} else {
				s.tid++
			}
			if s.Get(tid) == nil {
				return tid
			}
		}
	}
	return -1
}

func (s *searches) Insert(tor *ID, cb CallBack) (tid int16, sr *search) {
	if tid = s.nextTID(); tid != -1 {
		sr = newSearch(tor, cb)
		s.ss[tid] = sr
	}
	return
}

func (s *searches) Remove(tid int16) {
	delete(s.ss, tid)
}

func (s *searches) Map(f func(int16, *search) bool) {
	for tid, sr := range s.ss {
		if f(tid, sr) == false {
			return
		}
	}
}
