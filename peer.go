package dht

import (
	"net"
)

type Peer struct {
	addr *net.TCPAddr
}

func NewPeer(addr *net.TCPAddr) *Peer {
	return &Peer{
		addr: addr,
	}
}
