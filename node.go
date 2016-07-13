package dht

import (
	"fmt"
	"net"
)

// Node represent a dht peer
type Node struct {
	id   *ID
	ip   string
	port int
	addr *net.UDPAddr
}

// NewNode returns a node
func NewNode(id *ID, ip string, port int) *Node {
	return &Node{
		id:   id,
		ip:   ip,
		port: port,
	}
}

func NewNode2(id *ID, addr *net.UDPAddr) *Node {
	return &Node{
		id:   id,
		addr: addr,
	}
}

// Addr returns udb address
func (n *Node) Addr() *net.UDPAddr {
	return n.addr
}

// IsGood returns false if (now - n.time) > 15s
func (n *Node) IsGood() bool {
	return false
}

func (n *Node) String() string {
	return fmt.Sprintf("%v %v", n.id, n.addr)
}
