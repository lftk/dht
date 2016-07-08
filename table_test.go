package dht

import (
	"math/rand"
	"testing"
)

func Benchmark_Table_Append(b *testing.B) {
	total := rand.Intn(10000)
	t := NewTable(NewRandomID())
	for i := 0; i < total; i++ {
		n := NewNode(NewRandomID(), "127.0.0.1", 1234)
		t.Append(n)
	}
	var numsBucket, numsNode int
	t.Map(func(b *Bucket) bool {
		numsBucket++
		numsNode += b.Count()
		return true
	})
	b.Logf("total:%d\tbucket:%d\tnode:%d\n", total, numsBucket, numsNode)
}

func Test_Table_Lookup(t *testing.T) {
	t1 := NewTable(NewRandomID())
	for i := 0; i < rand.Intn(1000); i++ {
		n := NewNode(NewRandomID(), "127.0.0.1", 1234)
		t1.Append(n)
	}
	id1 := NewRandomID()
	n1 := t1.Lookup(id1)
	t.Log(id1)
	for _, n := range n1 {
		t.Log(n.id)
	}
}
