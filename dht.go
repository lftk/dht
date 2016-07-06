package dht

import (
	"fmt"
	"net"
	"time"
)

// DHT server
type DHT struct {
	cfg     *Config
	exit    chan bool
	conn    *net.UDPConn
	handler Handler
}

// NewDHT returns DHT
func NewDHT(cfg *Config) *DHT {
	return &DHT{
		cfg:  cfg,
		exit: make(chan bool),
	}
}

// Run dht server
func (d *DHT) Run(addr string, handler Handler) (err error) {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return
	}
	defer conn.Close()
	d.conn = conn.(*net.UDPConn)
	d.handler = handler
	d.loop()
	return
}

// Exit dht server
func (d *DHT) Exit() {
	close(d.exit)
}

func (d *DHT) loop() {
	fmt.Println(d.conn.LocalAddr())

	countdown := 10
	cleanup := time.Tick(30 * time.Second)
	for {
		select {
		case <-d.exit:
			return
		case <-cleanup:
			fmt.Println("cleanup...")
			if countdown--; countdown < 0 {
				d.Exit()
			}
		}
	}
}

// Ping a node
func (d *DHT) Ping(n *Node) error {
	return nil
}

// FindNode method
func (d *DHT) FindNode(id ID) error {
	return nil
}

// GetPeers method
func (d *DHT) GetPeers(id ID) error {
	return nil
}

// AnnouncePeer method
func (d *DHT) AnnouncePeer() error {
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
