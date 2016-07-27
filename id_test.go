package dht

import (
	"math/rand"
	"testing"
	"time"
)

func Test_NewID(t *testing.T) {
	s := "cf3865c78df296e8b5e8770663f4bce416e33335"
	id, err := ResolveID(s)
	if err != nil {
		t.Fatal(err)
	}
	if id.String() != s {
		t.Fatal(s, id)
	}
	id2, err := NewID(id.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if id2.Compare(id) != 0 {
		t.Fatal(id2, id)
	}
}

func Benchmark_NewID(b *testing.B) {
	for i := 0; i < rand.Intn(10000); i++ {
		newRandomID()
	}
}

var idRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func newRandomID() *ID {
	id := new(ID)
	n, err := idRand.Read(id[:])
	if err != nil || n != IDLen {
		return ZeroID
	}
	return id
}
