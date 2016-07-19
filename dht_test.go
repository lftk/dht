package dht

import (
	"math"
	"testing"
)

func Test_TID(t *testing.T) {
	testTID(t, "ping", 0)
	testTID(t, "find_node", 0)
	testTID(t, "get_peers", 0)
	for i := uint16(0); i < math.MaxUint16; i++ {
		testTID(t, "announce_peer", i)
	}
}

func testTID(t *testing.T, q string, id uint16) {
	tid := encodeTID(q, id)
	q2, id2 := decodeTID(tid)
	if q2 != q || id2 != id {
		t.Fatal(q, id, q2, id2)
	}
}
