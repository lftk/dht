package dht

import (
	"testing"
)

func Test_packet_Marshal(t *testing.T) {
	p := packetData{
		T: "aa",
		Y: "q",
	}
	b, err := p.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))
}

func Test_packet_Unmarshal(t *testing.T) {
	b1 := []byte("d1:eli201e23:A Generic Error Ocurrede1:t2:aa1:y1:ee")
	testPacketUnmarshal(t, b1)

	b2 := []byte("d1:ad2:id20:abcdefghij0123456789e1:q4:ping1:t2:aa1:y1:qe")
	testPacketUnmarshal(t, b2)

	b3 := []byte("d1:ad2:id20:abcdefghij01234567896:target20:mnopqrstuvwxyz123456e1:q9:find_node1:t2:aa1:y1:qe")
	testPacketUnmarshal(t, b3)

	b4 := []byte("d1:ad2:id20:abcdefghij01234567899:info_hash20:mnopqrstuvwxyz123456e1:q9:get_peers1:t2:aa1:y1:qe")
	testPacketUnmarshal(t, b4)
}

func testPacketUnmarshal(t *testing.T, b []byte) {
	var p packetData
	err := p.Unmarshal(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	/*
		if string(b2) != string(b) {
			t.Fatal(string(b2), string(b))
		}
	*/
}
