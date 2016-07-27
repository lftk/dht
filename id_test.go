package dht

import (
	"testing"
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
