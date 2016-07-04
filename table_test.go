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
	for e := t.buckets.Front(); e != nil; e = e.Next() {
		numsBucket++
		numsNode += e.Value.(*Bucket).Count()
	}
	b.Logf("total:%d\tbucket:%d\tnode:%d\n", total, numsBucket, numsNode)
	//b.Error(t)
}
