package dht

import (
	"fmt"
	"net"
	"time"
)

// Node represent a dht peer
type Node struct {
	id   *ID
	addr *net.UDPAddr
	time time.Time
}

// NewNode returns a node
func NewNode(id *ID, addr *net.UDPAddr) *Node {
	return &Node{
		id:   id,
		addr: addr,
		time: time.Now(),
	}
}

// ID returns id
func (n *Node) ID() *ID {
	return n.id
}

// Addr returns udb address
func (n *Node) Addr() *net.UDPAddr {
	return n.addr
}

// IsGood returns false if (now - n.time) > 15s
func (n *Node) IsGood() bool {
	return false
}

// Update node lasttime
func (n *Node) Update() {
	n.time = time.Now()
}

func (n *Node) String() string {
	return fmt.Sprintf("%v %v", n.id, n.addr)
}
