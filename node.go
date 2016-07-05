package dht

import (
	"fmt"
	"net"
)

// Node represent a dht peer
type Node struct {
	id   ID
	ip   string
	port int
}

// NewNode returns a node
func NewNode(id ID, ip string, port int) *Node {
	return &Node{
		id:   id,
		ip:   ip,
		port: port,
	}
}

// Addr returns udb address
func (n *Node) Addr() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   nil,
		Port: n.port,
	}
}

func (n *Node) String() string {
	return fmt.Sprintf("%v %s %d", n.id, n.ip, n.port)
}
