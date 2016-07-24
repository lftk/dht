package dht

import (
	"time"
)

type storage struct {
	id *ID
	ps map[string]time.Time
}

func newStorage(id *ID) *storage {
	return &storage{
		id: id,
		ps: make(map[string]time.Time),
	}
}

func (s *storage) Count() int {
	return len(s.ps)
}

func (s *storage) Insert(peer string) {
	s.ps[peer] = time.Now()
}

func (s *storage) Remove(peer string) {
	delete(s.ps, peer)
}

func (s *storage) Map(f func(p string, t time.Time) bool) {
	for peer, time := range s.ps {
		if f(peer, time) == false {
			return
		}
	}
}

type storages struct {
	ss map[ID]*storage
}

func newStorages() *storages {
	return &storages{
		ss: make(map[ID]*storage),
	}
}

func (s *storages) Count() int {
	return len(s.ss)
}

func (s *storages) Find(id *ID) *storage {
	if st, ok := s.ss[*id]; ok {
		return st
	}
	return nil
}

func (s *storages) Get(id *ID) (st *storage) {
	if st = s.Find(id); st == nil {
		st = newStorage(id)
		s.ss[*id] = st
	}
	return
}
