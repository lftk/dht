package dht

import (
	"net"
	"testing"
)

func Test_ResolvePeer(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:6881")
	if err != nil {
		t.Error(err)
	} else {
		peer := createPeer(addr.IP, addr.Port)
		ip, port := ResolvePeer(peer)
		if ip != "127.0.0.1" || port != 6881 {
			t.Error(addr, ip, port)
		}
	}
}
