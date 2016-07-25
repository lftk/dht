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

func (s *storage) ID() *ID {
	return s.id
}

func (s *storage) Count() int {
	return len(s.ps)
}

func (s *storage) Insert(peer []byte) {
	s.ps[string(peer)] = time.Now()
}

func (s *storage) Remove(peer []byte) {
	delete(s.ps, string(peer))
}

func (s *storage) Map(f func(p []byte, t time.Time) bool) {
	for peer, time := range s.ps {
		if f([]byte(peer), time) == false {
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

func (s *storages) Remove(id *ID) {
	delete(s.ss, *id)
}

func (s *storages) Map(f func(st *storage) bool) {
	for _, st := range s.ss {
		if f(st) == false {
			return
		}
	}
}
