package dht

import (
	"math/rand"
	"testing"
)

func Benchmark_NewNode(b *testing.B) {
	for i := 0; i < rand.Intn(10000); i++ {
		NewNode(ZeroID, nil)
	}
}
