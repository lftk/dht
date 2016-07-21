package dht

import "testing"

func Test_packet_Marshal(t *testing.T) {
}

func Test_packet_Unmarshal(t *testing.T) {
	/*
		b1 := []byte("d1:eli201e23:A Generic Error Ocurrede1:t2:aa1:y1:ee")
		b2 := []byte("d1:ad2:id20:abcdefghij0123456789e1:q4:ping1:t2:aa1:y1:qe")
		b3 := []byte("d1:ad2:id20:abcdefghij01234567896:target20:mnopqrstuvwxyz123456e1:q9:find_node1:t2:aa1:y1:qe")
		b4 := []byte("d1:ad2:id20:abcdefghij01234567899:info_hash20:mnopqrstuvwxyz123456e1:q9:get_peers1:t2:aa1:y1:qe")
	*/
}

func Test_newQueryMessage(t *testing.T) {
	/*
		id := NewRandomID()
		data := map[string]interface{}{"id": id.Bytes()}
		msg := NewQueryMessage("pn", "ping", data)
		b, err := bencode.EncodeBytes(msg)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(b))
	*/
}
