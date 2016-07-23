package dht

import (
	"fmt"
	"net"
	"time"
)

// Node represent a dht peer
type Node struct {
	id     *ID
	addr   *net.UDPAddr
	time   time.Time
	pinged int
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

// Time returns time
func (n *Node) Time() time.Time {
	return n.time
}

// Update node lasttime
func (n *Node) Update() {
	n.time = time.Now()
	n.pinged = 0
}

func (n *Node) String() string {
	return fmt.Sprintf("%v %v", n.id, n.addr)
}
