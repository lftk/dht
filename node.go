package dht

import "fmt"

type node struct {
	id   hash
	ip   string
	port int
}

func newNode(id hash, ip string, port int) *node {
	return &node{
		id:   id,
		ip:   ip,
		port: port,
	}
}

func (n *node) ping() {
}

func (n *node) String() string {
	return fmt.Sprintf("%v %s %d", n.id, n.ip, n.port)
}
