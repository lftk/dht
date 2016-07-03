package dht

import (
	"math/rand"
	"testing"
)

func Benchmark_table_append(b *testing.B) {
	total := rand.Intn(10000)
	t := newTable(newRandomHashWithoutError())
	for i := 0; i < total; i++ {
		n := newNode(newRandomHashWithoutError(), "127.0.0.1", 1234)
		t.append(n)
	}
	var numsBucket, numsNode int
	for e := t.buckets.Front(); e != nil; e = e.Next() {
		numsBucket++
		numsNode += e.Value.(*bucket).count()
	}
	b.Logf("total:%d\tbucket:%d\tnode:%d\n", total, numsBucket, numsNode)
}

func newRandomHashWithoutError() hash {
	h, _ := newRandomHash()
	return h
}
