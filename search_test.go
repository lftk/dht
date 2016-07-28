package dht

import (
	"math"
	"testing"
)

func Test_search(t *testing.T) {
	s := newSearches()
	for i := int16(0); i <= math.MaxInt16; i++ {
		if id, _ := s.Insert(nil, nil); id != i {
			t.Error(id, i)
			break
		}
		if i == math.MaxInt16 {
			break
		}
	}
}
