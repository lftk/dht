package dht

import (
	"testing"
)

func Test_search(t *testing.T) {
	s := newSearches()
	for i := int16(0); i < 1000; i++ {
		s.Insert(i, nil, func(tid int16, peer []byte) {
			//t.Log(tid)
		})
	}
	s.Map(func(sr *search) bool {
		sr.cb(sr.tid, nil)
		return true
	})
}
