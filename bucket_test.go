package dht

import (
	"math/rand"
	"testing"
)

func Test_Bucket(t *testing.T) {
	b := NewBucket(ZeroID, 8)
	for i := 0; i < rand.Intn(100); i++ {
		id := newRandomID()
		b.Insert(id, nil)
	}
	b.Map(func(n *Node) bool {
		if b.Find(n.id) != n {
			t.Error(n)
		}
		return true
	})
}
