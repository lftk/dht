package dht

import (
	"math"
	"testing"
)

func Test_search(t *testing.T) {
	s := newSearches()
	for i := int16(0); i <= math.MaxInt16; i++ {
		var idPtr *int16
		id, _ := s.Insert(nil, func(tid int16, peer []byte) {
			if idPtr == nil || *idPtr != tid {
				t.Error(tid)
			}
		})
		idPtr = &id
		if i == math.MaxInt16 {
			break
		}
	}
	s.Map(func(tid int16, sr *search) bool {
		sr.cb(tid, nil)
		return true
	})
}
