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
	for {
		p, addr, err := d.recvPacket()
		if err != nil {
			//
		}
		fmt.Println(p, addr, err)

		select {
		case <-d.exit:
			return
		}
	}
}

func (d *DHT) bootstrap() {
	d.FindNode(d.route.id)
}

// Ping a node
func (d *DHT) Ping(n *Node) error {
	data := map[string]interface{}{
		"id": d.id.Bytes(),
	}
	msg := newQueryMessage("ping", data)
	return sendMessage(d.conn, n.addr, msg)
}

// FindNode method
func (d *DHT) FindNode(id ID) error {
	data := map[string]interface{}{
		"id":     d.id.Bytes(),
		"target": id.Bytes(),
	}
	msg := newQueryMessage("find_node", data)
	for _, n := range d.route.Lookup(id) {
		sendMessage(d.conn, n.addr, msg)
	}
	return nil
}

// GetPeers method
func (d *DHT) GetPeers(id ID) error {
	data := map[string]interface{}{
		"id":        d.id.Bytes(),
		"info_hash": id.Bytes(),
	}
	msg := newQueryMessage("get_peers", data)
	for _, n := range d.route.Lookup(id) {
		sendMessage(d.conn, n.addr, msg)
	}
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
	var p packetData
	d.sendPacket(n, &p)
	return
}

func (d *DHT) replyGetPeers(n *Node) error {
	var p packetData
	d.sendPacket(n, &p)
	return nil
}

func (d *DHT) replyAnnouncePeer(n *Node) error {
	var p packetData
	d.sendPacket(n, &p)
	return nil
}

func (d *DHT) sendPacket(n *Node, p *packetData) (err error) {
	if b, err := p.Marshal(); err == nil {
		_, err = d.conn.WriteToUDP(b, n.Addr())
	}
	return
}

func (d *DHT) recvPacket() (p *packetData, addr *net.UDPAddr, err error) {
	_, addr, err = d.conn.ReadFromUDP(d.buf)
	if err != nil {
		return
	}
	p = new(packetData)
	err = p.Unmarshal(d.buf)
	return
}
