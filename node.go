package dht

import "fmt"

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

func (n *Node) String() string {
	return fmt.Sprintf("%v %s %d", n.id, n.ip, n.port)
}
