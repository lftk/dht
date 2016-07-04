package dht

import (
	"math/rand"
	"testing"
)

func Test_NewID(t *testing.T) {
	h := "cf3865c78df296e8b5e8770663f4bce416e33335"
	id, err := NewID([]byte(h))
	if err != nil {
		t.Fatal(err)
	}
	if id.String() != h {
		t.Fatal(id[:])
		t.Fatal(h, id)
	}
}

func Benchmark_NewRandomID(b *testing.B) {
	for i := 0; i < rand.Intn(10000); i++ {
		NewRandomID()
	}
}
