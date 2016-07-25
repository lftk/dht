package dht

import (
	"container/list"
	"math"
	"time"
)

// CallBack function
type CallBack func(tid int16, peer []byte)

type node struct {
	id   *ID
	time time.Time
}

type search struct {
	tor   *ID
	port  int
	cb    CallBack
	nodes *list.List
}

func newSearch(tor *ID, port int, cb CallBack) *search {
	return &search{
		tor:   tor,
		port:  port,
		cb:    cb,
		nodes: list.New(),
	}
}

func (s *search) Count() int {
	return s.nodes.Len()
}

func (s *search) Remove(id *ID) {
	s.handle(func(e *list.Element) bool {
		if e.Value.(*node).id.Compare(id) == 0 {
			s.nodes.Remove(e)
			return false
		}
		return true
	})
}

func (s *search) Map(f func(n *node) bool) {
	s.handle(func(e *list.Element) bool {
		return f(e.Value.(*node))
	})
}

func (s *search) handle(f func(e *list.Element) bool) {
	for e := s.nodes.Front(); e != nil; e = e.Next() {
		if f(e) == false {
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
	if n := s.Count(); n < math.MaxInt16 {
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

func (s *searches) Insert(tor *ID, port int, cb CallBack) (tid int16, sr *search) {
	if tid = s.nextTID(); tid != -1 {
		sr = newSearch(tor, port, cb)
		s.ss[tid] = sr
	}
	return
}

func (s *searches) Map(f func(tid int16, sr *search) bool) {
	for tid, sr := range s.ss {
		if f(tid, sr) == false {
			return
		}
	}
}
