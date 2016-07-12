package dht

import (
	"fmt"
	"net"
	"time"
)

// DHT server
type DHT struct {
	cfg     *Config
	id      ID
	buf     []byte
	exit    chan bool
	conn    *net.UDPConn
	route   *Table
	handler Handler
}

// NewDHT returns DHT
func NewDHT(cfg *Config) *DHT {
	return &DHT{
		cfg:   cfg,
		buf:   make([]byte, 4096),
		exit:  make(chan bool),
		route: NewTable(NewRandomID()),
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
	go d.loopPacket()

	d.bootstrap()

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

func (d *DHT) loopPacket() {
	buf := make([]byte, 4096)
	for {
		n, addr, err := d.conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		fmt.Println(addr, string(buf[:n]))

		select {
		case <-d.exit:
			return
		}
	}
}

func (d *DHT) bootstrap() {
	d.findNode(d.route.id)
}

func (d *DHT) ping(n *Node) {
	data := map[string]interface{}{
		"id": d.id.Bytes(),
	}
	msg := newQueryMessage("pn", "ping", data)
	d.sendMessage([]*Node{n}, msg)
}

func (d *DHT) findNode(id *ID) {
	data := map[string]interface{}{
		"id":     d.id.Bytes(),
		"target": id.Bytes(),
	}
	msg := newQueryMessage("fn", "find_node", data)
	d.sendMessage(d.route.Lookup(id), msg)
}

func (d *DHT) getPeers(id *ID) {
	data := map[string]interface{}{
		"id":        d.id.Bytes(),
		"info_hash": id.Bytes(),
	}
	msg := newQueryMessage("gp", "get_peers", data)
	d.sendMessage(d.route.Lookup(id), msg)
}

func (d *DHT) sendMessage(nodes []*Node, data interface{}) {
	addrs := make([]*net.UDPAddr, 0, len(nodes))
	for i, node := range nodes {
		addrs[i] = node.Addr()
	}
	sendUDPMessage(d.conn, addrs, data)
}

func (d *DHT) announcePeer() {
}

func (d *DHT) replyPing(n *Node) {
}

func (d *DHT) replyFindNode(n *Node) {
}

func (d *DHT) replyGetPeers(n *Node) {
}

func (d *DHT) replyAnnouncePeer(n *Node) {
}
