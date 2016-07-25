package dht

import (
	"container/list"
	"time"
)

// CallBack function
type CallBack func(tid int16, peer []byte)

type node struct {
	id   *ID
	time time.Time
}

type search struct {
	tid   int16
	tor   *ID
	cb    CallBack
	nodes *list.List
}

func newSearch(tid int16, tor *ID, cb CallBack) *search {
	return &search{
		tid:   tid,
		tor:   tor,
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
	ss map[int16]*search
}

func newSearches() *searches {
	return &searches{
		ss: make(map[int16]*search),
	}
}

func (s *searches) Find(tid int16) *search {
	if sr, ok := s.ss[tid]; ok {
		return sr
	}
	return nil
}

func (s *searches) Insert(tid int16, tor *ID, cb CallBack) *search {
	sr := newSearch(tid, tor, cb)
	s.ss[tid] = sr
	return sr
}

func (s *searches) Map(f func(s *search) bool) {
	for _, sr := range s.ss {
		if f(sr) == false {
			return
		}
	}
}
