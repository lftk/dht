package dht

import (
	"net"
)

// DHT server
type DHT struct {
	conn *net.UDPConn
}

func (d *DHT) ping() error {
	return nil
}

func (d *DHT) findNode() error {
	return nil
}

func (d *DHT) getPeers() error {
	return nil
}

func (d *DHT) announcePeer() error {
	return nil
}

func (d *DHT) replyPing(n *Node) (err error) {

	return
}

func (d *DHT) replyFindNode(n *Node) (err error) {
	var p packet
	d.sendPacket(n, &p)
	return
}

func (d *DHT) replyGetPeers(n *Node) error {
	var p packet
	d.sendPacket(n, &p)
	return nil
}

func (d *DHT) replyAnnouncePeer(n *Node) error {
	var p packet
	d.sendPacket(n, &p)
	return nil
}

func (d *DHT) sendPacket(n *Node, p *packet) (err error) {
	if b, err := p.Marshal(); err == nil {
		_, err = d.conn.WriteToUDP(b, n.Addr())
	}
	return
}
